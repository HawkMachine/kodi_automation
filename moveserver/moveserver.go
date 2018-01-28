package moveserver

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/HawkMachine/kodi_automation/platform"
	"github.com/HawkMachine/kodi_automation/utils/collections"

	kd "github.com/HawkMachine/kodi_go_api/v6/kodi"
	tr "github.com/HawkMachine/transmission_go_api"
)

// Returns a list of paths for all files and firectories in the source directory.
func directoryListing(dirname string, levels int, dirsOnly bool) ([]string, error) {
	res := []string{}
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if dirsOnly && !info.IsDir() {
			return nil
		}
		tmp := path
		for i := 0; i < levels; i++ {
			tmp = filepath.Dir(tmp)
			if tmp == dirname {
				res = append(res, path)
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	log.Printf("directoryListing %s %v", dirname, res)
	return res, nil
}

type MoveRequest struct {
	Path string
	To   string
}

type MoveListenerRequest struct {
	Request MoveRequest
}

func moveListener(s *MoveServer, ch chan MoveListenerRequest) {
	for {
		req := <-ch
		log.Printf("Received move requests %v", req)

		bytes, err := exec.Command("mv", req.Request.Path, req.Request.To).CombinedOutput()
		log.Printf("Move result: err: %v; output: %s", err, string(bytes))
		s.SetPathMoveResult(req.Request.Path, err, string(bytes))
	}
}

func updateDiskStats(s *MoveServer) {
	regex := regexp.MustCompile(`([^\s]+)\s+([^\s]+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)%`)
	// Results in MB
	bytes, err := exec.Command("df", "-B", "1", "--output=target,fstype,size,used,avail,pcent").Output()
	if err != nil {
		s.Log("UpdateDiskStats", fmt.Sprintf("df failed: %v", err))
		s.setDiskStats(nil)
	} else {
		ds := []DiskStats{}
		output := string(bytes)
		for _, line := range strings.Split(string(output), "\n")[1:] {
			grs := regex.FindStringSubmatch(line)
			if grs == nil {
				continue
			}
			size, err := strconv.ParseInt(grs[3], 10, 64)
			if err != nil {
				continue
			}
			used, err := strconv.ParseInt(grs[4], 10, 64)
			if err != nil {
				continue
			}
			avail, err := strconv.ParseInt(grs[5], 10, 64)
			if err != nil {
				continue
			}
			pFull, err := strconv.Atoi(grs[6])
			if err != nil {
				continue
			}
			if grs != nil {
				ds = append(ds, DiskStats{
					Path:        grs[1],
					FsType:      grs[2],
					Size:        size,
					Used:        used,
					Avail:       avail,
					PercentFull: pFull,
				})
			}
		}
		log.Print("Disk stats:")
		for _, x := range ds {
			log.Printf("   %#v", x)
		}
		s.setDiskStats(ds)
	}
}

func diskStatsUpdater(s *MoveServer, d time.Duration) {
	for {
		updateDiskStats(s)
		time.Sleep(d)
	}
}

func updateCache(s *MoveServer) {
	log.Printf("Updating cached info.")

	// List the files in the transmission directory.
	sourceDirListing, err := directoryListing(s.sourceDir, 1, false)
	if err != nil {
		s.Log("UpdateCache", fmt.Sprintf("Listing source target error: %v", err))
		sourceDirListing = nil
	}

	// List series from the target directories to generate the suggestions.
	var suggestionsList []string
	for seriesTarget := range s.seriesTargets {
		sl, err := getSeriesTargetListing(seriesTarget)
		if err != nil {
			fmt.Printf("Listing series target error: %v", err)
			continue
		}
		suggestionsList = append(suggestionsList, sl...)
	}

	// List all torrents from Transmission.
	torrentsList, err := s.t.ListAll()
	if err != nil {
		s.Log("UpdateCache", fmt.Sprintf("Getting torrents info error: %v", err))
		torrentsList = nil
	}

	// Get transmission info.

	s.setCachedInfo(sourceDirListing, torrentsList, suggestionsList)
}

func cacheUpdater(s *MoveServer, d time.Duration) {
	for {
		updateCache(s)
		time.Sleep(d)
	}
}

func (s *MoveServer) UpdateCacheAsync() {
	go func() {
		updateCache(s)
	}()
}

func (s *MoveServer) UpdateDiskStatsAsync() {
	go func() {
		updateDiskStats(s)
	}()
}

// LogMessage encapsulates message with a type and time.
type LogMessage struct {
	Type string
	Msg  string
	T    time.Time
}

// PathInfo has all the information kept on Server about paths in the Download directory.
type PathMoveInfo struct {
	Moving          bool
	Target          string
	LastError       error
	LastErrorOutput string
}

// Information about tranmission files.
type PathInfo struct {
	Name           string
	Path           string // Present if found on disk.
	AllowMove      bool
	AllowAssistant bool
	MoveInfo       PathMoveInfo
	Torrent        *tr.Torrent // Present if found in torrent.
	MoveTo         string      // Path where this should be moved, can be empty
}

type DiskStats struct {
	Path        string
	Size        int64
	Used        int64
	Avail       int64
	PercentFull int
	FsType      string
}

type MoveServerConfig struct {
	SourceDir         string   `json:"source_dir"`
	MoviesTargets     []string `json:"movies_targets"`
	SeriesTargets     []string `json:"series_targets"`
	MaxMvCommands     int      `json:"max_mv_commands"`
	MvBufferSize      int      `json:"mv_buffer_size"`
	DefaultMoveTarget string   `json:"default_move_target"`
}

type MoveServer struct {
	p *platform.Platform
	t *tr.Transmission
	k *kd.Kodi

	// Directory to scan
	sourceDir string

	// Targets for movies
	moviesTargets map[string]bool

	// Target for series and directory listing.
	seriesTargets map[string]bool

	// Move targets
	moveTargets        map[string]bool
	moveTargets_sorted []string

	// old path info kept for history - sorted by when things were moved
	pathInfoHistory []*PathInfo

	// Information for path that disappeared from the directory without being
	// moved with this moveserver.
	pathInfoDisappeared []*PathInfo

	// Path info is kept separately.
	pathInfo        map[string]*PathInfo
	cacheRefreshed  time.Time
	refreshDuration time.Duration

	// Default path where torrents are moved to.
	defaultMoveTarget string

	// Disk stats
	diskStats []DiskStats

	// Move channels.
	moveChannel chan MoveListenerRequest

	// Messages
	messages []*LogMessage

	lock         sync.Mutex
	messagesLock sync.Mutex

	// Move assistant.
	Assistant *Assistant
}

func New(p *platform.Platform, c MoveServerConfig) (*MoveServer, error) {
	t, _ := tr.New(
		p.Config.Transmission.Address,
		p.Config.Transmission.Username,
		p.Config.Transmission.Password)
	k := kd.New(
		p.Config.Kodi.Address,
		p.Config.Kodi.Username,
		p.Config.Kodi.Password)
	s := &MoveServer{
		p:                 p,
		t:                 t,
		k:                 k,
		sourceDir:         c.SourceDir,
		moviesTargets:     collections.NewStringsSet(c.MoviesTargets),
		seriesTargets:     collections.NewStringsSet(c.SeriesTargets),
		pathInfoHistory:   []*PathInfo{},
		pathInfo:          map[string]*PathInfo{},
		refreshDuration:   5 * time.Minute,
		cacheRefreshed:    time.Now(),
		moveChannel:       make(chan MoveListenerRequest, c.MvBufferSize),
		messages:          []*LogMessage{},
		lock:              sync.Mutex{},
		messagesLock:      sync.Mutex{},
		defaultMoveTarget: c.DefaultMoveTarget,
	}

	for i := 0; i < c.MaxMvCommands; i++ {
		go moveListener(s, s.moveChannel)
	}
	go cacheUpdater(s, 5*time.Minute)
	go diskStatsUpdater(s, 5*time.Minute)

	// If we have a target path of not it makes sense to have an assitant.  It
	// will start torrents and move them to their destination when they're
	// finished.
	s.Assistant = newAssistant(s)
	s.Assistant.Enable()
	s.Log("moveserver", fmt.Sprintf("Assistant created, default target path %s", s.defaultMoveTarget))

	return s, nil
}

func (s *MoveServer) GetCacheRefreshed() time.Time {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.cacheRefreshed
}

func (s *MoveServer) GetMoveTargets() []string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.moveTargets_sorted
}

func (s *MoveServer) GetPathInfoAndPathInfoHistory() (map[string]*PathInfo, []*PathInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// TODO: we should create a copy here.
	return s.pathInfo, s.pathInfoHistory
}

func (s *MoveServer) GetMvBuffSizeAndElems() (int, int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return cap(s.moveChannel), len(s.moveChannel)
}

func (s *MoveServer) GetDiskStats() []DiskStats {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.diskStats
}

func (s *MoveServer) GetMessages() []*LogMessage {
	s.messagesLock.Lock()
	defer s.messagesLock.Unlock()

	return s.messages
}

func (s *MoveServer) Log(tp, msg string) {
	s.messagesLock.Lock()
	defer s.messagesLock.Unlock()

	m := &LogMessage{
		Type: tp,
		T:    time.Now(),
		Msg:  msg,
	}
	log.Printf("%s: %s", tp, msg)
	if len(s.messages) >= 5000 {
		s.messages = s.messages[:5000]
	}
	s.messages = append([]*LogMessage{m}, s.messages...)
}

func (s *MoveServer) SetMovePath(name, move_to string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	pi, ok := s.pathInfo[name]
	if !ok {
		return fmt.Errorf("Item %s not found.", name)
	}
	pi.MoveTo = move_to
	return nil
}

func (s *MoveServer) Move(name string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	pi, ok := s.pathInfo[name]
	if !ok {
		return fmt.Errorf("Item %s not found.", name)
	}
	return s.moveLocked(pi)
}

func (s *MoveServer) validateMovePathInfo(pi *PathInfo) error {
	if pi.Path == "" {
		return fmt.Errorf("No path associated with %s, likely only in transmission.", pi.Name)
	}
	if !pi.AllowMove {
		return fmt.Errorf("Moving the path is not allowed")
	}
	if pi.MoveInfo.Moving {
		return fmt.Errorf("Requested move path %s is currently in move to %s", pi.MoveInfo.Target)
	}
	if _, err := os.Stat(pi.Path); err != nil && os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *MoveServer) validateMoveTargetPath(path string) error {
	if !s.moveTargets[path] {
		return fmt.Errorf("Requested move target %s not on move targets list: %v", path, s.moveTargets)
	}
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *MoveServer) moveLocked(pi *PathInfo) error {
	// Source path verification
	if err := s.validateMovePathInfo(pi); err != nil {
		pi.MoveInfo.LastError = err
		return err
	}

	// Target path verification
	if err := s.validateMoveTargetPath(pi.MoveTo); err != nil {
		pi.MoveInfo.LastError = err
		return err
	}

	// Queue verification.
	if len(s.moveChannel) == cap(s.moveChannel) {
		return fmt.Errorf("Mv requests buffer buffer is full.")
	}

	// Actually making a move.
	target := filepath.Join(pi.MoveTo, filepath.Base(pi.Path))
	pi.MoveInfo.Moving = true
	pi.MoveInfo.Target = target

	s.moveChannel <- MoveListenerRequest{
		Request: MoveRequest{
			Path: pi.Path,
			To:   target,
		},
	}
	return nil
}

func (s *MoveServer) SetPathMoveResult(path string, err error, output string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	name := filepath.Base(path)
	pi, ok := s.pathInfo[name]
	if !ok {
		return fmt.Errorf("path not found")
	}

	pi.AllowMove = false
	pi.MoveInfo = PathMoveInfo{
		Moving:          false,
		Target:          pi.MoveInfo.Target,
		LastError:       err,
		LastErrorOutput: output,
	}
	if err == nil {
		// Successful move.
		delete(s.pathInfo, name)
		s.pathInfoHistory = append(s.pathInfoHistory, pi)
		s.Log("MoveResult", fmt.Sprintf("Successfully moved %s to %s", pi.Name, pi.MoveInfo.Target))
	} else {
		// Unsuccessful move.
		pi.MoveInfo.Target = ""
	}
	return nil
}

func (s *MoveServer) loadTorrentsInfo() ([]*tr.Torrent, error) {
	// In the future I need more to ask both - transmission and scan local disk.
	return s.t.ListAll()
}

func getSeriesTargetListing(seriesTarget string) ([]string, error) {
	listing, err := directoryListing(seriesTarget, 2, true)
	if err != nil {
		return nil, err
	}
	return listing, nil
}

// setCachedInfo updates cahed infor on MoveServer. paths is a list of paths in
// the source dir, ntis is the new transmission info, nstl is the new
// series target listing.
func (s *MoveServer) setCachedInfo(sourceDirListing []string, torrenstListing []*tr.Torrent, suggestionsListing []string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// If any of the needed values are nil, that means there was an error. For
	// now, we do not do anything with that data.
	if sourceDirListing == nil {
		s.Log("SetCachedInfo", fmt.Sprintf("got nil: paths=%v, ntis=%v, nstl=%s", sourceDirListing, torrenstListing, suggestionsListing))
		return
	}

	oldPathInfo := s.pathInfo

	// Create new path info for paths that were found on disk.
	newPathInfo := map[string]*PathInfo{}
	for _, path := range sourceDirListing {
		name := filepath.Base(path)
		newPathInfo[name] = &PathInfo{
			Name:           name,
			Path:           path,
			AllowAssistant: true,
			MoveTo:         s.defaultMoveTarget,
		}
	}

	// Add new path info from the transmission data that was found.
	for _, t := range torrenstListing {
		pi, ok := newPathInfo[t.Name]
		if !ok {
			newPathInfo[t.Name] = &PathInfo{
				Name:           t.Name,
				Torrent:        t,
				AllowAssistant: true,
				MoveTo:         s.defaultMoveTarget,
			}
		} else {
			// Update the torrent field for this path.
			pi.Torrent = t
		}
	}

	// Copy fields from old path info to newly created path info structures.
	for _, opi := range oldPathInfo {
		if pi, ok := newPathInfo[opi.Name]; ok {
			pi.AllowAssistant = opi.AllowAssistant
			pi.AllowMove = opi.AllowMove
			pi.MoveInfo = opi.MoveInfo
			pi.MoveTo = opi.MoveTo
		} else {
			s.pathInfoDisappeared = append(s.pathInfoDisappeared, opi)
		}
	}

	// Update AllowMove
	for _, pi := range newPathInfo {
		pi.AllowMove = pi.Path != ""
		if pi.Torrent != nil {
			pi.AllowMove = pi.Torrent.PercentDone == 1.0 && !pi.MoveInfo.Moving
		}
	}

	s.cacheRefreshed = time.Now()
	s.pathInfo = newPathInfo

	moveTargets := map[string]bool{}
	for _, seriesListingPath := range suggestionsListing {
		moveTargets[seriesListingPath] = true
	}
	for seriesTarget := range s.seriesTargets {
		moveTargets[seriesTarget] = true
	}
	seriesTargets_sorted := []string{}
	for target := range moveTargets {
		seriesTargets_sorted = append(seriesTargets_sorted, target)
	}
	sort.Strings(seriesTargets_sorted)

	moveTargets_sorted := []string{}
	for moviesTarget := range s.moviesTargets {
		moveTargets[moviesTarget] = true
		moveTargets_sorted = append(moveTargets_sorted, moviesTarget)
	}
	moveTargets_sorted = append(moveTargets_sorted, seriesTargets_sorted...)

	s.moveTargets_sorted = moveTargets_sorted
	s.moveTargets = moveTargets
}

func (s *MoveServer) setDiskStats(nds []DiskStats) {
	s.lock.Lock()
	defer s.lock.Unlock()

	diskStats := []DiskStats{}
	for _, ds := range nds {
		for moveTarget := range s.moveTargets {
			if strings.HasPrefix(moveTarget, ds.Path) {
				diskStats = append(diskStats, ds)
				break
			}
		}
	}
	s.diskStats = diskStats
}

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

	"github.com/HawkMachine/kodi_automation/utils/collections"
	"github.com/HawkMachine/transmission_go_api"
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
	bytes, err := exec.Command("df", "-B", "1000000", "--output=target,fstype,size,used,avail,pcent").Output()
	if err != nil {
		log.Printf("df: %v", err)
		s.setDiskStats(nil)
	} else {
		ds := []DiskStats{}
		output := string(bytes)
		for _, line := range strings.Split(string(output), "\n")[1:] {
			grs := regex.FindStringSubmatch(line)
			if grs == nil {
				continue
			}
			size, err := strconv.Atoi(grs[3])
			if err != nil {
				continue
			}
			used, err := strconv.Atoi(grs[4])
			if err != nil {
				continue
			}
			avail, err := strconv.Atoi(grs[5])
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
	tranmissionLst, err := directoryListing(s.sourceDir, 1, false)
	if err != nil {
		log.Printf("Listing source target error: %v\n", err)
		tranmissionLst = nil
	}

	// List series from the target directories to generate the suggestions.
	var seriesListing []string
	for seriesTarget := range s.seriesTargets {
		sl, err := getSeriesTargetListing(seriesTarget)
		if err != nil {
			log.Printf("Listing series target error: %v\n", err)
			continue
		}
		seriesListing = append(seriesListing, sl...)
	}

	// List all torrents from Transmission.
	tis, err := s.t.ListAll()
	if err != nil {
		log.Printf("Getting torrents info error: %v\n", err)
		tis = nil
	}

	s.setCachedInfo(tranmissionLst, tis, seriesListing)
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

// msgTimestamp encapsulates message with a type and time.
type msgTimestamp struct {
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
	Name      string
	Path      string // Present if found on disk.
	AllowMove bool
	MoveInfo  PathMoveInfo
	Torrent   *transmission_go_api.Torrent // Present if found in torrent.
	// TODO: add field for files found on disk.
}

type DiskStats struct {
	Path        string
	Size        int
	Used        int
	Avail       int
	PercentFull int
	FsType      string
}

type MoveServer struct {
	t *transmission_go_api.Transmission

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
	// move with this moveserver.
	pathInfoDisappeared []*PathInfo

	// Path info is kept separately.
	pathInfo        map[string]*PathInfo
	cacheRefreshed  time.Time
	refreshDuration time.Duration

	// Disk stats
	diskStats []DiskStats

	// chennels
	moveChannel chan MoveListenerRequest

	// Messages
	messages []*msgTimestamp

	lock sync.Mutex
}

func New(sourceDir string, moviesTarget []string, seriesTargets []string,
	maxMvComands int, mvBufferSize int,
	address, username, password string) (*MoveServer, error) {
	t, _ := transmission_go_api.New(address, username, password)
	s := &MoveServer{
		t:               t,
		sourceDir:       sourceDir,
		moviesTargets:   collections.NewStringsSet(moviesTarget),
		seriesTargets:   collections.NewStringsSet(seriesTargets),
		pathInfoHistory: []*PathInfo{},
		pathInfo:        map[string]*PathInfo{},
		refreshDuration: 5 * time.Minute,
		cacheRefreshed:  time.Now(),
		moveChannel:     make(chan MoveListenerRequest, mvBufferSize),
		messages:        []*msgTimestamp{},
		lock:            sync.Mutex{},
	}

	for i := 0; i < maxMvComands; i++ {
		go moveListener(s, s.moveChannel)
	}
	go cacheUpdater(s, 5*time.Minute)
	go diskStatsUpdater(s, 5*time.Minute)

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

func (s *MoveServer) Move(path string, to string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// ------------ Path verification ---------------
	pi, ok := s.pathInfo[path]
	if !ok {
		return fmt.Errorf("Requested move path %s not found in our data bank", path)
	}
	if filepath.Dir(path) != s.sourceDir {
		return fmt.Errorf("Trying to move file outside of %s", s.sourceDir)
	}
	if pi.MoveInfo.Moving {
		return fmt.Errorf("Requested move path %s is currently in move to %s", pi.MoveInfo.Target)
	}
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		pi.MoveInfo.LastError = err
		s.pathInfo[path] = pi
		return err
	}

	// Target path verification
	if !s.moveTargets[to] {
		return fmt.Errorf("Requested target unrecognized")
	}
	if _, err := os.Stat(to); err != nil && os.IsNotExist(err) {
		pi.MoveInfo.LastError = err
		s.pathInfo[path] = pi
		return err
	}

	// Queue verification.
	if len(s.moveChannel) == cap(s.moveChannel) {
		return fmt.Errorf("Mv requests buffer buffer is full.")
	}

	// Actually making a move.
	target := filepath.Join(to, filepath.Base(path))
	pi.MoveInfo.Moving = true
	pi.MoveInfo.Target = target

	s.moveChannel <- MoveListenerRequest{
		Request: MoveRequest{
			Path: path,
			To:   target,
		},
	}
	s.pathInfo[path] = pi
	return nil
}

func (s *MoveServer) SetPathMoveResult(path string, err error, output string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	pi, ok := s.pathInfo[path]
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
		delete(s.pathInfo, path)
		s.pathInfoHistory = append(s.pathInfoHistory, pi)
	} else {
		// Unsuccessful move.
		pi.MoveInfo.Target = ""
		s.pathInfo[path] = pi
	}
	return nil
}

func (s *MoveServer) loadTorrentsInfo() ([]*transmission_go_api.Torrent, error) {
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
func (s *MoveServer) setCachedInfo(paths []string, ntis []*transmission_go_api.Torrent, nstl []string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// If any of the needed values are nil, that means there was an error. For
	// now, we do not do anything with that data.
	if paths == nil || ntis == nil || nstl == nil {
		m := &msgTimestamp{
			Type: "Set cached info",
			T:    time.Now(),
			Msg:  fmt.Sprintf("got nil: paths=%v, ntis=%v, nstl=%s", paths, ntis, nstl),
		}
		log.Printf("setCachedInfo: %v", m)
		s.messages = append(s.messages, m)
		return
	}

	oldPathInfo := s.pathInfo

	// TODO: currently I set AllowMove to true. I need to check the torrent state
	// for that and also allow move if transmission doesn't know about it.

	// Create new path info for paths that were found on disk.
	newPathInfo := map[string]*PathInfo{}
	for _, path := range paths {
		name := filepath.Base(path)
		newPathInfo[name] = &PathInfo{
			Name:      name,
			Path:      path,
			AllowMove: true,
		}
	}

	// Add new path info from the transmission data that was found.
	for _, t := range ntis {
		pi, ok := newPathInfo[t.Name]
		if !ok {
			newPathInfo[t.Name] = &PathInfo{
				Name:      t.Name,
				AllowMove: true,
				Torrent:   t,
			}
		} else {
			// Try to join path info from torrent info.
			pi.Torrent = t
		}
	}

	// Copy old move info to new data.
	for _, opi := range oldPathInfo {
		// TODO: keep the old move info in some sort of history.
		if pi, ok := newPathInfo[opi.Path]; ok {
			pi.MoveInfo = opi.MoveInfo
		} else {
			s.pathInfoDisappeared = append(s.pathInfoDisappeared, opi)
		}
	}

	// Update AllowMove
	//for _, path := range paths {
	//	pi := newPathInfo[path]
	//	// log.Printf("%v", pi)
	//	if pi.Torrent != nil {
	//		pi.AllowMove = pi.Torrent.IsFinished && !pi.MoveInfo.Moving
	//	}
	//	newPathInfo[path] = pi
	//}

	s.cacheRefreshed = time.Now()
	s.pathInfo = newPathInfo

	moveTargets := map[string]bool{}
	for _, seriesListingPath := range nstl {
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

package moveserver

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/HawkMachine/kodi_automation/transmission"
	"github.com/HawkMachine/kodi_automation/utils/collections"
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
	regex := regexp.MustCompile(`([^\s]+)\s+([^\s]+)\s+([^\s]+)\s+([^\s]+)\s+([^\s]+)%\s+(.*)`)
	bytes, err := exec.Command("df", "-h").Output()
	if err != nil {
		log.Printf("df -h: %v", err)
		s.setDiskStats(nil)
	} else {
		ds := []DiskStats{}
		output := string(bytes)
		for _, line := range strings.Split(string(output), "\n")[1:] {
			grs := regex.FindStringSubmatch(line)
			if grs == nil {
				continue
			}
			if grs != nil {
				ds = append(ds, DiskStats{
					Path:       grs[6],
					Filesystem: grs[1],
					Size:       grs[2],
					Used:       grs[3],
					Avail:      grs[4],
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

	paths, err := directoryListing(s.sourceDir, 1, false)
	if err != nil {
		log.Printf("Listing source target error: %v\n", err)
		paths = nil
	}

	var seriesListing []string
	for seriesTarget := range s.seriesTargets {
		sl, err := getSeriesTargetListing(seriesTarget)
		if err != nil {
			log.Printf("Listing series target error: %v\n", err)
			continue
		}
		seriesListing = append(seriesListing, sl...)
	}

	tis, err := s.transmissionCli.GetTorrents(s.sourceDir)
	if err != nil {
		log.Printf("Getting torrents info error: %v\n", err)
		tis = nil
	}

	s.setCachedInfo(paths, tis, seriesListing)
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

// PathInfo has all the information kept on Server about paths in the Download directory.
type PathMoveInfo struct {
	Moving          bool
	Target          string
	LastError       error
	LastErrorOutput string
}

type PathInfo struct {
	Name        string
	Path        string
	AllowMove   bool
	MoveInfo    PathMoveInfo
	TorrentInfo transmission.TorrentInfo
}

type DiskStats struct {
	Path       string
	Filesystem string
	Size       string
	Used       string
	Avail      string
}

type MoveServer struct {
	transmissionCli *transmission.Transmission

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
	pathInfoHistory []PathInfo

	// Path info is kept separately.
	pathInfo        map[string]PathInfo
	cacheRefreshed  time.Time
	refreshDuration time.Duration

	// Disk stats
	diskStats []DiskStats

	// chennels
	moveChannel chan MoveListenerRequest

	lock sync.Mutex
}

func New(sourceDir string, moviesTarget []string, seriesTargets []string, maxMvComands int, mvBufferSize int) (*MoveServer, error) {
	s := &MoveServer{
		transmissionCli: transmission.New("", 0, "", ""),
		sourceDir:       sourceDir,
		moviesTargets:   collections.NewStringsSet(moviesTarget),
		seriesTargets:   collections.NewStringsSet(seriesTargets),
		pathInfoHistory: []PathInfo{},
		pathInfo:        map[string]PathInfo{},
		refreshDuration: 5 * time.Minute,
		cacheRefreshed:  time.Now(),
		moveChannel:     make(chan MoveListenerRequest, mvBufferSize),
		lock:            sync.Mutex{},
	}

	for i := 0; i < maxMvComands; i++ {
		go moveListener(s, s.moveChannel)
	}
	go cacheUpdater(s, 5*time.Minute)
	go diskStatsUpdater(s, 5*time.Minute)

	return s, nil
}

func (s *MoveServer) GetPathInfo() map[string]PathInfo {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pathInfo
}

func (s *MoveServer) GetPathInfoHistory() []PathInfo {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pathInfoHistory
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

func (s *MoveServer) GetPathInfoAndPathInfoHistory() (map[string]PathInfo, []PathInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()

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

func (s *MoveServer) loadTorrentsInfo() ([]transmission.TorrentInfo, error) {
	// In the future I need more to ask both - transmission and scan local disk.
	return s.transmissionCli.GetTorrents(s.sourceDir)
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
func (s *MoveServer) setCachedInfo(paths []string, ntis []transmission.TorrentInfo, nstl []string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Set new info.
	oldPathInfo := s.pathInfo
	newPathInfo := map[string]PathInfo{}
	for _, path := range paths {
		newPathInfo[path] = PathInfo{
			Name:      filepath.Base(path),
			Path:      path,
			AllowMove: true,
			MoveInfo:  PathMoveInfo{},
		}
	}
	// Copy MoveInfo and TorrentInfo from old data.
	for _, opi := range oldPathInfo {
		pi, ok := newPathInfo[opi.Path]
		if ok {
			pi.MoveInfo = opi.MoveInfo

			// Copy old torrent info only if the new torrent info is nil (there was an
			// error while getting torrents info) or that path is currently moved
			// (torrent may have been deleted from transmission.
			var ti transmission.TorrentInfo
			if ntis == nil || pi.MoveInfo.Moving {
				ti = pi.TorrentInfo
			}
			pi.TorrentInfo = ti

			newPathInfo[opi.Path] = pi
		}
	}
	// Update path info with torrents info.
	for _, nti := range ntis {
		path := filepath.Join(nti.DownloadDir, nti.Name)
		pi, ok := newPathInfo[path]
		if ok {
			pi.TorrentInfo = nti
			newPathInfo[path] = pi
		}
	}
	// Update AllowMove
	for _, path := range paths {
		pi := newPathInfo[path]
		pi.AllowMove = pi.TorrentInfo.IsFinished && !pi.MoveInfo.Moving
		newPathInfo[path] = pi
	}

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

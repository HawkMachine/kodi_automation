package moveserver

import (
	"fmt"
	"kodi_automation/transmission"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	//"strconv"
	"strings"
	"sync"
	"time"
)

func directoryListing(dirname string, levels int, dirs_only bool) ([]string, error) {
	res := []string{}
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if dirs_only && !info.IsDir() {
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

		bytes, err := exec.Command("mv", req.Request.Path, req.Request.To).Output()
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
			//size, err1 := strconv.ParseInt(grs[2], 10, 64)
			//used, err2 := strconv.ParseInt(grs[3], 10, 64)
			//avail, err3 := strconv.ParseInt(grs[4], 10, 64)
			//if err1 != nil || err2 != nil || err3 != nil {
			//	log.Printf("Atoi errors: %v, %v, %v", err1, err2, err3)
			//	continue
			//}
			if grs != nil {
				ds = append(ds, DiskStats{
					Path:       grs[6],
					Filesystem: grs[1],
					Size:       grs[2], //size,
					Used:       grs[3], // used,
					Avail:      grs[4], //avail,
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
	// regex := regexp.MustCompile(`([^\s]+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)%\s+(.*)`)
	for {
		updateDiskStats(s)
		time.Sleep(d)
	}
}

func updateCache(s *MoveServer) {
	log.Printf("Updating cached info.")
	paths, err := directoryListing(s.sourceDir, 1, false)
	sl, slErr := getSeriesTargetListing(s.seriesTarget)
	tis, _ := s.transmissionCli.GetTorrents(s.sourceDir)

	if err != nil || slErr != nil {
		log.Printf(fmt.Sprintf("Update cached info errors: %v; %v", err, slErr))
	} else {
		s.setCachedInfo(paths, tis, sl)
	}
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
	moviesTarget string

	// Target for series and directory listing.
	seriesTarget        string
	seriesTargetListing []string

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

func New(sourceDir string, moviesTarget string, seriesTarget string, maxMvComands int, mvBufferSize int) (*MoveServer, error) {
	s := &MoveServer{
		transmissionCli: transmission.New("", 0, "", ""),
		sourceDir:       sourceDir,
		moviesTarget:    moviesTarget,
		seriesTarget:    seriesTarget,
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

func (s *MoveServer) GetSeriesTarget() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.seriesTarget
}

func (s *MoveServer) GetMoviesTarget() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.moviesTarget
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

func (s *MoveServer) GetSeriesTargetListing() []string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.seriesTargetListing
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

func (s *MoveServer) getTargetForSeriesMove(path, to string) (string, error) {
	if filepath.IsAbs(to) {
		return "", fmt.Errorf("Target path is an absolute path, expected relative: %s", to)
	}
	target, err := filepath.Abs(filepath.Join(s.seriesTarget, to, filepath.Base(path)))
	if err != nil {
		return "", err
	}
	if target[:len(s.seriesTarget)] != s.seriesTarget {
		return "", fmt.Errorf("Tried to move path outside series dir, to %s", target)
	}
	return target, nil
}

func (s *MoveServer) getTargetForMovieMove(path string) (string, error) {
	target, err := filepath.Abs(filepath.Join(s.moviesTarget, filepath.Base(path)))
	if err != nil {
		return "", err
	}
	if filepath.Dir(target) != s.moviesTarget {
		return "", fmt.Errorf("Tried to move outside movies dir: %s", target)
	}
	return target, nil
}

func (s *MoveServer) Move(path string, to string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// ------------ Path verification ---------------
	if filepath.Dir(path) != s.sourceDir {
		return fmt.Errorf("Trying to move file outside of %s", s.sourceDir)
	}
	pi, ok := s.pathInfo[path]
	if !ok {
		return fmt.Errorf("Requested move path %s not found in our data bank", path)
	}
	if pi.MoveInfo.Moving {
		return fmt.Errorf("Requested move path %s is currently in move to %s", pi.MoveInfo.Target)
	}
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		pi.MoveInfo.LastError = err
		s.pathInfo[path] = pi
		return err
	}

	// ------------ Target verification ---------------
	var target string
	var err error
	if to == "movies" {
		target, err = s.getTargetForMovieMove(path)
	} else {
		target, err = s.getTargetForSeriesMove(path, to)
	}
	if err != nil {
		pi.MoveInfo.LastError = err
		s.pathInfo[path] = pi
		return err
	}
	if len(s.moveChannel) == cap(s.moveChannel) {
		return fmt.Errorf("Mv requests buffer buffer is full")
	}

	// Actually making a move.
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
}

func (s *MoveServer) setDiskStats(nds []DiskStats) {
	s.lock.Lock()
	defer s.lock.Unlock()

	diskStats := []DiskStats{}
	for _, ds := range nds {
		if strings.HasPrefix(s.seriesTarget, ds.Path) || strings.HasPrefix(s.moviesTarget, ds.Path) {
			diskStats = append(diskStats, ds)
		}
	}
	s.diskStats = diskStats
}

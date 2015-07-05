package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"kodi_automation/transmission"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var (
	MV_BUFFER_SIZE  = flag.Int("mv_buffer_size", 5, "size of the mv request buffer")
	MAX_MV_COMMANDS = flag.Int("max_mv_commands", 5, "max mv commands running in parallel")
	PORT            = flag.Int("port", 8080, "port to use")
	SOURCE_DIR      = flag.String("source_dir", "", "directory to scan")
	MOVIES_TARGET   = flag.String("movies_dir", "", "where to move movies")
	SERIES_TARGET   = flag.String("series_dir", "", "where to move series")
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

func MakeHandler(h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h(w, r)
		d := time.Since(start)
		footer := fmt.Sprintf("Elapsed time: %v", d)
		fmt.Fprintf(w, "\n\n--------\n\n")
		fmt.Fprintf(w, footer)
	}
}

// -------------------- Server -------------------

type MoveRequest struct {
	Path string
	To   string
}

type MoveCancelRequest struct {
}

type MoveStatusRequest struct {
}

type MoveReply struct {
	Path  string
	Reply string
}

type MoveListenerRequest struct {
	Type          string // request, cancel or status or quit
	Request       MoveRequest
	CancelRequest MoveCancelRequest
	StatusRequest MoveStatusRequest
}

func MoveListener(s *Server, ch chan MoveListenerRequest) {
	for {
		req := <-ch
		log.Printf("MoveListener received requests %v", req)
		switch req.Type {
		case "move":
			err := exec.Command("mv", req.Request.Path, req.Request.To).Run()
			if err != nil {
				s.SetPathMoveError(req.Request.Path, err)
			} else {
				s.SetPathMoved(req.Request.Path)
			}
		case "quit":
			break
		}
	}
}

func CacheUpdater(s *Server, d time.Duration) {
	for {
		time.Sleep(d)
		s.UpdateCache()
	}
}

// PathInfo has all the information kept on Server about paths in the Download directory.
type PathMoveInfo struct {
	Moving bool
	Error  error
	Target string
	WorkId int
}

type PathInfo struct {
	Name        string
	Path        string
	AllowMove   bool
	MoveInfo    PathMoveInfo
	TorrentInfo transmission.TorrentInfo
}
type PathInfoSlice []PathInfo

// Make slice of PathInfo sortable.
func (pi PathInfoSlice) Len() int {
	return len(pi)
}
func (pi PathInfoSlice) Less(i, j int) bool {
	return pi[i].Path < pi[j].Path
}
func (pi PathInfoSlice) Swap(i, j int) {
	pi[i], pi[j] = pi[j], pi[i]
}

type Server struct {
	templates       map[string]string
	parsedTemplates map[string]*template.Template
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

	// chennels
	moveChannel chan MoveListenerRequest

	lock sync.Mutex
}

func (s *Server) Init(sourceDir string, moviesTarget string, seriesTarget string, port int, maxMvComands int, mvBufferSize int) error {
	s.transmissionCli = transmission.New("", 0, "", "")
	s.sourceDir = sourceDir
	s.moviesTarget = moviesTarget
	s.seriesTarget = seriesTarget
	s.pathInfoHistory = []PathInfo{}
	s.pathInfo = nil
	s.refreshDuration = 5 * time.Minute
	s.cacheRefreshed = time.Now()

	s.moveChannel = make(chan MoveListenerRequest, mvBufferSize)
	for i := 0; i < maxMvComands; i++ {
		go MoveListener(s, s.moveChannel)
	}
	go CacheUpdater(s, time.Minute)

	s.lock = sync.Mutex{}

	// Load templates
	s.parsedTemplates = make(map[string]*template.Template)
	for name, path := range s.templates {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		t := template.New(name)
		t, err = t.Parse(string(data))
		if err != nil {
			return err
		}
		s.parsedTemplates[name] = t
	}
	return nil
}

func (s *Server) GetSeriesTarget() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.seriesTarget
}

func (s *Server) GetMoviesTarget() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.moviesTarget
}

func (s *Server) GetPathInfo() map[string]PathInfo {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pathInfo
}

func (s *Server) GetPathInfoHistory() []PathInfo {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pathInfoHistory
}

func (s *Server) GetCacheRefreshed() time.Time {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.cacheRefreshed
}

func (s *Server) GetSeriesTargetListing() []string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.seriesTargetListing
}

func (s *Server) GetPathInfoAndPathInfoHistory() (map[string]PathInfo, []PathInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pathInfo, s.pathInfoHistory
}

func (s *Server) GetMvBuffSizeAndElems() (int, int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return cap(s.moveChannel), len(s.moveChannel)
}

func (s *Server) GetTemplate(name string) (*template.Template, error) {
	t, ok := s.parsedTemplates[name]
	if !ok {
		return nil, fmt.Errorf("Unknown template %s", name)
	}
	return t, nil
}

func (s *Server) getTargetForSeriesMove(path, to string) (string, error) {
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

func (s *Server) getTargetForMovieMove(path string) (string, error) {
	target, err := filepath.Abs(filepath.Join(s.moviesTarget, filepath.Base(path)))
	if err != nil {
		return "", err
	}
	if filepath.Dir(target) != s.moviesTarget {
		return "", fmt.Errorf("Tried to move outside movies dir: %s", target)
	}
	return target, nil
}

func (s *Server) Move(path string, to string) error {
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

	// ------------ Target verification ---------------
	var target string
	var err error
	if to == "movies" {
		target, err = s.getTargetForMovieMove(path)
	} else {
		target, err = s.getTargetForSeriesMove(path, to)
	}
	if err != nil {
		pi.MoveInfo.Error = err
		s.pathInfo[path] = pi
		return err
	}
	if len(s.moveChannel) == cap(s.moveChannel) {
		return fmt.Errorf("Mv requests buffer buffer is full")
	}

	// Actually making a move.
	pi.MoveInfo = PathMoveInfo{
		Moving: true,
		Target: target,
		Error:  pi.MoveInfo.Error, // remember last error
	}
	s.moveChannel <- MoveListenerRequest{
		Type: "move",
		Request: MoveRequest{
			Path: path,
			To:   target,
		},
	}
	s.pathInfo[path] = pi
	return nil
}

func (s *Server) SetPathMoved(path string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	pi, ok := s.pathInfo[path]
	if !ok {
		return fmt.Errorf("path not found")
	}
	pi.MoveInfo.Moving = false
	pi.MoveInfo.Error = nil
	delete(s.pathInfo, path)
	s.pathInfoHistory = append(s.pathInfoHistory, pi)
	return nil
}

func (s *Server) SetPathMoveError(path string, err error) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	pi, ok := s.pathInfo[path]
	if !ok {
		return fmt.Errorf("path not found")
	}
	pi.MoveInfo.Moving = false
	pi.MoveInfo.Error = err
	s.pathInfo[path] = pi
	return nil
}

func (s *Server) loadTorrentsInfo() ([]transmission.TorrentInfo, error) {
	// In the future I need more to ask both - transmission and scan local disk.
	return s.transmissionCli.GetTorrents(s.sourceDir)
}

func (s *Server) updatePathInfoCache() error {
	// Try updating cache if it's the first time
	log.Printf("Updating path info cache")
	// Retrieve current data.
	tis, _ := s.loadTorrentsInfo()
	paths, err := directoryListing(s.sourceDir, 1, false)
	if err != nil {
		log.Printf("Directory listing error: %v", err)
		return err
	}

	// Update server with new information.
	pathInfoHistory := s.pathInfo
	s.cacheRefreshed = time.Now()
	s.pathInfo = map[string]PathInfo{}
	for _, path := range paths {
		if pi, ok := pathInfoHistory[path]; ok {
			s.pathInfo[path] = pi
		} else {
			s.pathInfo[path] = PathInfo{
				Path:      path,
				AllowMove: true,
				MoveInfo: PathMoveInfo{
					Moving: false,
				},
			}
		}
	}

	// Update torrent info for paths.
	if tis != nil {
		for _, ti := range tis {
			for i := range s.pathInfo {
				if s.pathInfo[i].Path == filepath.Join(ti.DownloadDir, ti.Name) {
					pi := s.pathInfo[i]
					pi.TorrentInfo = ti
					pi.AllowMove = ti.IsFinished
					s.pathInfo[i] = pi
				}
			}
		}
	}
	return nil
}

func (s *Server) updateSeriesTargetListing() error {
	listing, err := directoryListing(s.seriesTarget, 2, true)
	if err != nil {
		return err
	}
	s.seriesTargetListing = listing
	return nil
}

func (s *Server) UpdateCache() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	piErr := s.updatePathInfoCache()
	slErr := s.updateSeriesTargetListing()

	if piErr != nil || slErr != nil {
		return fmt.Errorf(fmt.Sprintf("UpdateCache errors: %v %v", piErr, slErr))
	}
	return nil
}

func (s *Server) UpdateSeriesTargetListing() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.updateSeriesTargetListing()
}

func (s *Server) MakeHandler(h func(*Server, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h(s, w, r)
		d := time.Since(start)
		footer := fmt.Sprintf("Elapsed time: %v", d)
		fmt.Fprintf(w, "\n\n--------\n\n")
		fmt.Fprintf(w, footer)
	}
}

// -------------------- Page handlers -------------------

func movePostHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	log.Printf("Received move POST request %v", r)
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	form := r.Form
	messageType := form.Get("type")
	switch messageType {
	case "":
		err = fmt.Errorf("Submitted POST is missing type value")
	case "move":
		err = s.Move(form.Get("path"), form.Get("target"))
	default:
		err = fmt.Errorf("Submitted POST has unrecognized type value: %s", messageType)
	}
	//	if err != nil {
	//		http.Error(w, err.Error(), http.StatusInternalServerError)
	//	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func updateCacheHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	log.Printf("Received cache update POST request %v", r)
	s.UpdateCache()
	http.Redirect(w, r, "/", http.StatusFound)
}

func pathInfoPageHandler(s *Server, w http.ResponseWriter, r *http.Request) {
	var err, postErr error
	t, err := s.GetTemplate("torrents_page")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	// ---- prepare pathInfo and pathInfoHistory
	pathInfo, pathInfoHistory := s.GetPathInfoAndPathInfoHistory()
	pathInfoList := PathInfoSlice{}
	for _, pi := range pathInfo {
		pathInfoList = append(pathInfoList, pi)
	}
	pathInfoHistoryList := PathInfoSlice{}
	for _, pi := range pathInfoHistory {
		pathInfoHistoryList = append(pathInfoHistoryList, pi)
	}

	// ---- prepare target listing
	seriesTargetListing := s.GetSeriesTargetListing()
	seriesTarget := s.GetSeriesTarget()
	seriesTargets := []string{}
	for _, sTarget := range seriesTargetListing {
		seriesTargets = append(seriesTargets, sTarget[len(seriesTarget)+1:])
	}

	mvBuffSize, mvBuffElems := s.GetMvBuffSizeAndElems()

	sort.Sort(pathInfoList)
	context := struct {
		PathInfo        []PathInfo
		PathInfoHistory []PathInfo
		SeriesTargets   []string
		Errors          []error
		CacheResfreshed time.Time
		MvBufferSize    int
		MvBufferElems   int
	}{
		PathInfo:        pathInfoList,
		PathInfoHistory: pathInfoHistoryList,
		SeriesTargets:   seriesTargets,
		Errors: []error{
			postErr,
		},
		CacheResfreshed: s.GetCacheRefreshed(),
		MvBufferSize:    mvBuffSize,
		MvBufferElems:   mvBuffElems,
	}
	err = t.ExecuteTemplate(w, "torrents_page", context)
	if err != nil {
		fmt.Fprintf(w, fmt.Sprintf("Error: %v", err))
		return
	}
}

func main() {
	flag.Parse()
	var err error
	if *SOURCE_DIR == "" || *MOVIES_TARGET == "" || *SERIES_TARGET == "" {
		log.Fatal("Missing source, series or movies target directory")
	}

	if *SOURCE_DIR, err = filepath.Abs(*SOURCE_DIR); err != nil {
		log.Fatal(err)
	}
	if *MOVIES_TARGET, err = filepath.Abs(*MOVIES_TARGET); err != nil {
		log.Fatal(err)
	}
	if *SERIES_TARGET, err = filepath.Abs(*SERIES_TARGET); err != nil {
		log.Fatal(err)
	}
	if *MAX_MV_COMMANDS <= 5 {
		*MAX_MV_COMMANDS = 5
	}
	if *MV_BUFFER_SIZE <= 5 {
		*MV_BUFFER_SIZE = 5
	}
	log.Printf("PORT             = %d", *PORT)
	log.Printf("SOURCE_DIR       = %s", *SOURCE_DIR)
	log.Printf("MOVIES_TARGET    = %s", *MOVIES_TARGET)
	log.Printf("SERIES_TARGET    = %s", *SERIES_TARGET)
	log.Printf("MAX_MV_COMMANDS  = %d", *MAX_MV_COMMANDS)
	log.Printf("MV_BUFFER_SIZE   = %d", *MV_BUFFER_SIZE)

	server := Server{
		templates: map[string]string{
			"torrents_page": "torrents_page.html",
		},
	}
	err = server.Init(*SOURCE_DIR, *MOVIES_TARGET, *SERIES_TARGET, *PORT, *MAX_MV_COMMANDS, *MV_BUFFER_SIZE)
	server.UpdateCache()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", server.MakeHandler(pathInfoPageHandler))
	http.HandleFunc("/move", server.MakeHandler(movePostHandler))
	http.HandleFunc("/cache/update", server.MakeHandler(updateCacheHandler))
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	bindAddr := fmt.Sprintf(":%d", *PORT)
	log.Printf("Bind address %s", bindAddr)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}

package main

import (
	"flag"
	"fmt"
	"html/template"
	"kodi_automation/moveserver"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	// "time"
)

var (
	MV_BUFFER_SIZE   = flag.Int("mv_buffer_size", 5, "size of the mv request buffer")
	MAX_MV_COMMANDS  = flag.Int("max_mv_commands", 5, "max mv commands running in parallel")
	PORT             = flag.Int("port", 8080, "port to use")
	SOURCE_DIR       = flag.String("source_dir", "", "directory to scan")
	MOVIES_TARGET    = flag.String("movies_dir", "", "where to move movies")
	SERIES_TARGET    = flag.String("series_dir", "", "where to move series")
	CUSTOM_LINKS     = flag.String("custom_links", "", "comma-delimited list of <link name>:<url>")
	TRANSMISSION_URL = flag.String("transmission_url", "", "transmission URL")
)

// -------------------- Server -------------------

type PathInfoSlice []moveserver.PathInfo

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

type MyHttpServer struct {
	parsedTemplates map[string]*template.Template
	moveServer      *moveserver.MoveServer
	links           map[string]string
	transmissionURL string
}

func New(moveServer *moveserver.MoveServer, transmissionURL string, templates map[string][]string, links map[string]string) (*MyHttpServer, error) {
	if moveServer == nil {
		return nil, fmt.Errorf("Move server cannot be null")
	}
	if links == nil {
		links = map[string]string{}
	}
	httpServer := &MyHttpServer{
		parsedTemplates: map[string]*template.Template{},
		links:           links,
		moveServer:      moveServer,
		transmissionURL: transmissionURL,
	}
	for name, paths := range templates {
		t := template.Must(template.New(name).ParseFiles(paths...))
		httpServer.parsedTemplates[name] = t
	}
	return httpServer, nil
}

func (s *MyHttpServer) GetTemplate(name string) (*template.Template, error) {
	t, ok := s.parsedTemplates[name]
	if !ok {
		return nil, fmt.Errorf("Unknown template %s", name)
	}
	return t, nil
}

func MakeHandler(s *MyHttpServer, h func(*MyHttpServer, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h(s, w, r)
	}
}

// -------------------- Page handlers -------------------

func movePostHandler(s *MyHttpServer, w http.ResponseWriter, r *http.Request) {
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
		err = s.moveServer.Move(form.Get("path"), form.Get("target"))
	default:
		err = fmt.Errorf("Submitted POST has unrecognized type value: %s", messageType)
	}
	//	if err != nil {
	//		http.Error(w, err.Error(), http.StatusInternalServerError)
	//	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func updateCacheHandler(s *MyHttpServer, w http.ResponseWriter, r *http.Request) {
	log.Printf("Received cache update POST request %v", r)
	s.moveServer.UpdateCache()
	http.Redirect(w, r, "/", http.StatusFound)
}

func transmissionPageHandler(s *MyHttpServer, w http.ResponseWriter, r *http.Request) {
	t, err := s.GetTemplate("transmission_page")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	context := struct {
		Errors          []error
		TransmissionURL string
		CustomLinks     map[string]string
	}{
		TransmissionURL: s.transmissionURL,
	}
	err = t.ExecuteTemplate(w, "base", context)
	if err != nil {
		fmt.Fprintf(w, fmt.Sprintf("Error: %v", err))
		return
	}
}

func pathInfoPageHandler(s *MyHttpServer, w http.ResponseWriter, r *http.Request) {
	var err, postErr error
	t, err := s.GetTemplate("torrents_page")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	// ---- prepare pathInfo and pathInfoHistory
	pathInfo, pathInfoHistory := s.moveServer.GetPathInfoAndPathInfoHistory()
	pathInfoList := PathInfoSlice{}
	for _, pi := range pathInfo {
		pathInfoList = append(pathInfoList, pi)
	}
	pathInfoHistoryList := PathInfoSlice{}
	for _, pi := range pathInfoHistory {
		pathInfoHistoryList = append(pathInfoHistoryList, pi)
	}

	// ---- prepare target listing
	seriesTargetListing := s.moveServer.GetSeriesTargetListing()
	seriesTarget := s.moveServer.GetSeriesTarget()
	seriesTargets := []string{}
	for _, sTarget := range seriesTargetListing {
		seriesTargets = append(seriesTargets, sTarget[len(seriesTarget)+1:])
	}

	mvBuffSize, mvBuffElems := s.moveServer.GetMvBuffSizeAndElems()

	formattedCacheRefreshed := s.moveServer.GetCacheRefreshed().Format("15:04 02-01-2006")

	sort.Sort(pathInfoList)
	context := struct {
		PathInfo        []moveserver.PathInfo
		PathInfoHistory []moveserver.PathInfo
		SeriesTargets   []string
		Errors          []error
		CacheResfreshed string
		MvBufferSize    int
		MvBufferElems   int
		CustomLinks     map[string]string
		DiskStats       []moveserver.DiskStats
	}{
		PathInfo:        pathInfoList,
		PathInfoHistory: pathInfoHistoryList,
		SeriesTargets:   seriesTargets,
		Errors: []error{
			postErr,
		},
		CacheResfreshed: formattedCacheRefreshed,
		MvBufferSize:    mvBuffSize,
		MvBufferElems:   mvBuffElems,
		CustomLinks:     s.links,
		DiskStats:       s.moveServer.GetDiskStats(),
	}
	err = t.ExecuteTemplate(w, "base", context)
	if err != nil {
		fmt.Fprintf(w, fmt.Sprintf("Error: %v", err))
		return
	}
}

func parseCustomLinksFlag(str string) map[string]string {
	res := map[string]string{}
	for _, item := range strings.Split(str, ",") {
		link_items := strings.SplitN(item, ":", 2)
		if len(link_items) < 2 {
			log.Printf("Wrong flag format")
			continue
		}
		res[link_items[0]] = link_items[1]
	}
	return res
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
	links := parseCustomLinksFlag(*CUSTOM_LINKS)
	log.Printf("PORT             = %d", *PORT)
	log.Printf("SOURCE_DIR       = %s", *SOURCE_DIR)
	log.Printf("MOVIES_TARGET    = %s", *MOVIES_TARGET)
	log.Printf("SERIES_TARGET    = %s", *SERIES_TARGET)
	log.Printf("MAX_MV_COMMANDS  = %d", *MAX_MV_COMMANDS)
	log.Printf("MV_BUFFER_SIZE   = %d", *MV_BUFFER_SIZE)
	log.Printf("TRANSMISSION URL = %s", *TRANSMISSION_URL)
	log.Printf("LINKS            = %v", links)

	moveServer, err := moveserver.New(*SOURCE_DIR, *MOVIES_TARGET, *SERIES_TARGET, *MAX_MV_COMMANDS, *MV_BUFFER_SIZE)
	if err != nil {
		panic(err)
	}
	server, err := New(
		moveServer,
		*TRANSMISSION_URL,
		map[string][]string{
			"torrents_page": {
				"templates/base.html",
				"templates/torrents_page.html",
			},
			"transmission_page": {
				"templates/base.html",
				"templates/transmission_page.html",
			},
		},
		links,
	)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", MakeHandler(server, pathInfoPageHandler))
	http.HandleFunc("/transmission", MakeHandler(server, transmissionPageHandler))
	http.HandleFunc("/move", MakeHandler(server, movePostHandler))
	http.HandleFunc("/cache/update", MakeHandler(server, updateCacheHandler))
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	bindAddr := fmt.Sprintf(":%d", *PORT)
	log.Printf("Bind address %s", bindAddr)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}

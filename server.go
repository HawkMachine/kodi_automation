package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/HawkMachine/kodi_automation/moveserver"
)

var (
	mvBufferSize      = flag.Int("mv_buffer_size", 5, "size of the mv request buffer")
	maxMvCommands     = flag.Int("max_mv_commands", 5, "max mv commands running in parallel")
	port              = flag.Int("port", 8080, "port to use")
	sourceDir         = flag.String("source_dir", "", "directory to scan")
	moviesTarget      = flag.String("movies_dir", "", "where to move movies")
	seriesTarget      = flag.String("series_dir", "", "where to move series")
	customLinks       = flag.String("links", "", "comma-delimited list of <link name>:<url>")
	customIframeLinks = flag.String("iframe_links", "", "comma-delimited list of <link name>:<url>")
	waitForIP         = flag.Int("wait_for_ip", 300, "Seconds to wait for IP address")

	templatesPath = flag.String("templates_path", "templates", "Path to server templates.")
	resourcesPath = flag.String("resources_path", "resources", "Path to server resources.")
)

// -------------------- Server -------------------

// PathInfoSlice is sortable list of moveserver.PathInfo.
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

type myHTTPServer struct {
	parsedTemplates map[string]*template.Template

	moveServer *moveserver.MoveServer

	links       map[string]string
	iframeLinks map[string]string
}

type basePageContext struct {
	Title          string
	IsMobile       bool
	ContentContext interface{}
	Links          map[string]string
	IframeLinks    map[string]string
	Errors         []string
}

// New creates new instance of myHTTPServer.
func newHTTPServer(moveServer *moveserver.MoveServer, templates map[string][]string, templatesPath string,
	links map[string]string, iframeLinks map[string]string) (*myHTTPServer, error) {

	if moveServer == nil {
		return nil, fmt.Errorf("Move server cannot be null")
	}
	if links == nil {
		links = map[string]string{}
	}
	if iframeLinks == nil {
		iframeLinks = map[string]string{}
	}
	httpServer := &myHTTPServer{
		parsedTemplates: map[string]*template.Template{},

		moveServer: moveServer,

		links:       links,
		iframeLinks: iframeLinks,
	}
	for name, relPaths := range templates {
		var paths []string
		for _, relPath := range relPaths {
			paths = append(paths, filepath.Join(templatesPath, relPath))
		}
		t := template.Must(template.New(name).ParseFiles(paths...))
		httpServer.parsedTemplates[name] = t
	}
	return httpServer, nil
}

func (s *myHTTPServer) GetTemplate(name string) (*template.Template, error) {
	t, ok := s.parsedTemplates[name]
	if !ok {
		return nil, fmt.Errorf("Unknown template %s", name)
	}
	return t, nil
}

func (s *myHTTPServer) GetBaseContext(title string, r *http.Request) basePageContext {
	return basePageContext{
		Title:       title,
		IsMobile:    s.IsMobile(r),
		Links:       s.links,
		IframeLinks: s.iframeLinks,
	}
}

func (s *myHTTPServer) GetIframeLink(name string) (string, bool) {
	v, ok := s.iframeLinks[name]
	return v, ok
}

func makeHandler(s *myHTTPServer, h func(*myHTTPServer, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h(s, w, r)
	}
}

func (s *myHTTPServer) IsMobile(r *http.Request) bool {
	// I'm so lazy...
	return strings.Contains(r.UserAgent(), "Nexus")
}

// GetLocalIP returns the non loopback local IP of the host.
// http://stackoverflow.com/a/31551220
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}

// -------------------- Page handlers -------------------

func movePostHandler(s *myHTTPServer, w http.ResponseWriter, r *http.Request) {
	log.Printf("Received move POST request", r)
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
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func updateCacheHandler(s *myHTTPServer, w http.ResponseWriter, r *http.Request) {
	log.Printf("Received cache update POST request %v", r)
	s.moveServer.UpdateCacheAsync()
	http.Redirect(w, r, "/", http.StatusFound)
}

func updateDiskStatsHandler(s *myHTTPServer, w http.ResponseWriter, r *http.Request) {
	log.Printf("Received disk stats POST request %v", r)
	s.moveServer.UpdateDiskStatsAsync()
	http.Redirect(w, r, "/", http.StatusFound)
}

func pathInfoPageHandler(s *myHTTPServer, w http.ResponseWriter, r *http.Request) {
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

	formattedCacheRefreshed := s.moveServer.GetCacheRefreshed().Format("15:04:05 02-01-2006")

	sort.Sort(pathInfoList)
	context := struct {
		PathInfo        []moveserver.PathInfo
		PathInfoHistory []moveserver.PathInfo
		SeriesTargets   []string
		Errors          []error
		CacheResfreshed string
		MvBufferSize    int
		MvBufferElems   int
		DiskStats       []moveserver.DiskStats
	}{
		PathInfo:        pathInfoList,
		PathInfoHistory: pathInfoHistoryList,
		SeriesTargets:   seriesTargets,
		CacheResfreshed: formattedCacheRefreshed,
		MvBufferSize:    mvBuffSize,
		MvBufferElems:   mvBuffElems,
		DiskStats:       s.moveServer.GetDiskStats(),
	}
	baseContext := s.GetBaseContext("", r)
	baseContext.ContentContext = context
	err = t.ExecuteTemplate(w, "base", baseContext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func wrapHandler(s *myHTTPServer, w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/wrap/"):]
	url, ok := s.GetIframeLink(name)
	fmt.Printf(url)
	if !ok {
		http.Error(w, fmt.Sprintf("Page %v not found", name), http.StatusNotFound)
	}

	t, err := s.GetTemplate("wrap_page")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	context := struct {
		WrapURL string
	}{
		WrapURL: url,
	}
	baseContext := s.GetBaseContext(strings.Title(name), r)
	baseContext.ContentContext = context
	err = t.ExecuteTemplate(w, "base", baseContext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseCustomLinksFlag(str string) map[string]string {
	res := map[string]string{}
	for _, item := range strings.Split(str, ",") {
		if item == "" {
			continue
		}
		linkItems := strings.SplitN(item, ":", 2)
		if len(linkItems) < 2 {
			log.Printf("Wrong flag format: ", item)
			continue
		}
		res[linkItems[0]] = linkItems[1]
	}
	return res
}

func replaceLocalHost(m map[string]string, ip string) map[string]string {
	n := map[string]string{}
	for k, v := range m {
		n[k] = strings.Replace(v, "localhost", ip, 1)
	}
	return n
}

func main() {
	flag.Parse()
	var err error
	if *sourceDir == "" || *moviesTarget == "" || *seriesTarget == "" {
		log.Fatal("Missing source, series or movies target directory")
	}

	if *sourceDir, err = filepath.Abs(*sourceDir); err != nil {
		log.Fatal(err)
	}
	if *moviesTarget, err = filepath.Abs(*moviesTarget); err != nil {
		log.Fatal(err)
	}
	if *seriesTarget, err = filepath.Abs(*seriesTarget); err != nil {
		log.Fatal(err)
	}
	if *maxMvCommands <= 5 {
		*maxMvCommands = 5
	}
	if *mvBufferSize <= 5 {
		*mvBufferSize = 5
	}
	log.Printf("PORT             = %d", *port)
	log.Printf("SOURCE_DIR       = %s", *sourceDir)
	log.Printf("MOVIES_TARGET    = %s", *moviesTarget)
	log.Printf("SERIES_TARGET    = %s", *seriesTarget)
	log.Printf("MAX_MV_COMMANDS  = %d", *maxMvCommands)
	log.Printf("MV_BUFFER_SIZE   = %d", *mvBufferSize)
	log.Printf("WAIT_FOR_IP      = %d", *waitForIP)
	log.Printf("LINKS            = %v", *customLinks)
	log.Printf("IFRAME LINKS     = %v", *customIframeLinks)

	var ip string
	start := time.Now()
	deadline := start.Add(time.Duration(*waitForIP) * time.Second)
	for {
		if time.Now().After(deadline) {
			log.Fatalf("Could not get local IP for %d seconds", *waitForIP)
		}
		ip, err = GetLocalIP()
		if err != nil {
			fmt.Printf("Got error while resolving IP: %v", err)
		} else if ip != "" {
			break
		} else {
			log.Printf("Resolved IP to empty string, %s left", deadline.Sub(time.Now()).String())
		}
		time.Sleep(time.Second * 5)
	}

	log.Printf("IP               = %s", ip)
	links := replaceLocalHost(parseCustomLinksFlag(*customLinks), ip)
	iframeLinks := replaceLocalHost(parseCustomLinksFlag(*customIframeLinks), ip)
	log.Printf("LINKS            = %v", links)
	log.Printf("IFRAME LINKS     = %v", iframeLinks)

	moveServer, err := moveserver.New(*sourceDir, *moviesTarget, *seriesTarget, *maxMvCommands, *mvBufferSize)
	if err != nil {
		panic(err)
	}
	templatesMap := map[string][]string{
		"torrents_page": {
			"base.html",
			"torrents_page.html",
		},
		"transmission_page": {
			"base.html",
			"transmission_page.html",
		},
		"wrap_page": {
			"base.html",
			"wrap_page.html",
		},
	}
	server, err := newHTTPServer(
		moveServer,
		templatesMap,
		*templatesPath,
		links,
		iframeLinks,
	)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", makeHandler(server, pathInfoPageHandler))
	// http.HandleFunc("/transmission", makeHandler(server, transmissionPageHandler))
	http.HandleFunc("/move", makeHandler(server, movePostHandler))
	http.HandleFunc("/update/cache", makeHandler(server, updateCacheHandler))
	http.HandleFunc("/update/disks", makeHandler(server, updateDiskStatsHandler))
	http.HandleFunc("/wrap/", makeHandler(server, wrapHandler))
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir(*resourcesPath))))

	bindAddr := fmt.Sprintf(":%d", *port)
	log.Printf("Bind address %s", bindAddr)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}

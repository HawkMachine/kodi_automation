package main

import (
	"flag"
	"fmt"
	"html/template"
	"kodi_automation/moveserver"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	MV_BUFFER_SIZE      = flag.Int("mv_buffer_size", 5, "size of the mv request buffer")
	MAX_MV_COMMANDS     = flag.Int("max_mv_commands", 5, "max mv commands running in parallel")
	PORT                = flag.Int("port", 8080, "port to use")
	SOURCE_DIR          = flag.String("source_dir", "", "directory to scan")
	MOVIES_TARGET       = flag.String("movies_dir", "", "where to move movies")
	SERIES_TARGET       = flag.String("series_dir", "", "where to move series")
	CUSTOM_LINKS        = flag.String("links", "", "comma-delimited list of <link name>:<url>")
	CUSTOM_IFRAME_LINKS = flag.String("iframe_links", "", "comma-delimited list of <link name>:<url>")
	WAIT_FOR_IP         = flag.Int("wait_for_ip", 300, "Seconds to wait for IP address")
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

	moveServer *moveserver.MoveServer

	links       map[string]string
	iframeLinks map[string]string
}

type BasePageContext struct {
	Title          string
	IsMobile       bool
	ContentContext interface{}
	Links          map[string]string
	IframeLinks    map[string]string
	Errors         []string
}

func New(moveServer *moveserver.MoveServer, templates map[string][]string,
	links map[string]string, iframeLinks map[string]string) (*MyHttpServer, error) {

	if moveServer == nil {
		return nil, fmt.Errorf("Move server cannot be null")
	}
	if links == nil {
		links = map[string]string{}
	}
	if iframeLinks == nil {
		iframeLinks = map[string]string{}
	}
	httpServer := &MyHttpServer{
		parsedTemplates: map[string]*template.Template{},

		moveServer: moveServer,

		links:       links,
		iframeLinks: iframeLinks,
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

func (s *MyHttpServer) GetBaseContext(title string, r *http.Request) BasePageContext {
	return BasePageContext{
		Title:       title,
		IsMobile:    s.IsMobile(r),
		Links:       s.links,
		IframeLinks: s.iframeLinks,
	}
}

func (s *MyHttpServer) GetIframeLink(name string) (string, bool) {
	v, ok := s.iframeLinks[name]
	return v, ok
}

func MakeHandler(s *MyHttpServer, h func(*MyHttpServer, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h(s, w, r)
	}
}

func (s *MyHttpServer) IsMobile(r *http.Request) bool {
	// I'm so lazy...
	return strings.Contains(r.UserAgent(), "Nexus")
}

// http://stackoverflow.com/a/31551220
// GetLocalIP returns the non loopback local IP of the host.
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

func movePostHandler(s *MyHttpServer, w http.ResponseWriter, r *http.Request) {
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

func updateCacheHandler(s *MyHttpServer, w http.ResponseWriter, r *http.Request) {
	log.Printf("Received cache update POST request %v", r)
	s.moveServer.UpdateCacheAsync()
	http.Redirect(w, r, "/", http.StatusFound)
}

func updateDiskStatsHandler(s *MyHttpServer, w http.ResponseWriter, r *http.Request) {
	log.Printf("Received disk stats POST request %v", r)
	s.moveServer.UpdateDiskStatsAsync()
	http.Redirect(w, r, "/", http.StatusFound)
}

func pathInfoPageHandler(s *MyHttpServer, w http.ResponseWriter, r *http.Request) {
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

func wrapHandler(s *MyHttpServer, w http.ResponseWriter, r *http.Request) {
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
		link_items := strings.SplitN(item, ":", 2)
		if len(link_items) < 2 {
			log.Printf("Wrong flag format: ", item)
			continue
		}
		res[link_items[0]] = link_items[1]
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
	log.Printf("WAIT FOR IP      = %d", *WAIT_FOR_IP)
	log.Printf("LINKS            = %v", *CUSTOM_LINKS)
	log.Printf("IFRAME LINKS     = %v", *CUSTOM_IFRAME_LINKS)

	var ip string
	start := time.Now()
	deadline := start.Add(time.Duration(*WAIT_FOR_IP) * time.Second)
	for {
		if time.Now().After(deadline) {
			log.Fatalf("Could not get local IP for %d seconds", *WAIT_FOR_IP)
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
	links := replaceLocalHost(parseCustomLinksFlag(*CUSTOM_LINKS), ip)
	iframeLinks := replaceLocalHost(parseCustomLinksFlag(*CUSTOM_IFRAME_LINKS), ip)
	log.Printf("LINKS            = %v", links)
	log.Printf("IFRAME LINKS     = %v", iframeLinks)

	moveServer, err := moveserver.New(*SOURCE_DIR, *MOVIES_TARGET, *SERIES_TARGET, *MAX_MV_COMMANDS, *MV_BUFFER_SIZE)
	if err != nil {
		panic(err)
	}
	server, err := New(
		moveServer,
		map[string][]string{
			"torrents_page": {
				"templates/base.html",
				"templates/torrents_page.html",
			},
			"transmission_page": {
				"templates/base.html",
				"templates/transmission_page.html",
			},
			"wrap_page": {
				"templates/base.html",
				"templates/wrap_page.html",
			},
		},
		links,
		iframeLinks,
	)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", MakeHandler(server, pathInfoPageHandler))
	// http.HandleFunc("/transmission", MakeHandler(server, transmissionPageHandler))
	http.HandleFunc("/move", MakeHandler(server, movePostHandler))
	http.HandleFunc("/update/cache", MakeHandler(server, updateCacheHandler))
	http.HandleFunc("/update/disks", MakeHandler(server, updateDiskStatsHandler))
	http.HandleFunc("/wrap/", MakeHandler(server, wrapHandler))
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	bindAddr := fmt.Sprintf(":%d", *PORT)
	log.Printf("Bind address %s", bindAddr)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}

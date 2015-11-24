package moveserverview

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/HawkMachine/kodi_automation/moveserver"
	"github.com/HawkMachine/kodi_automation/server"
)

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

// Implements server.View interface.
type MoveServerView struct {
	server_    server.HTTPServer
	moveServer *moveserver.MoveServer
}

func (msv *MoveServerView) GetName() string {
	return "moveserverview"
}

func (msv *MoveServerView) GetTemplates() map[string][]string {
	return map[string][]string{
		"torrents_page": []string{
			"base.html",
			"torrents_page.html",
		},
	}
}

func (msv *MoveServerView) GetHandlers() map[string]server.ViewHandle {
	return map[string]server.ViewHandle{
		"/":             server.NewViewHandle(msv.pathInfoPageHandler),
		"/move":         server.NewViewHandle(msv.movePostHandler),
		"/update/cache": server.NewViewHandle(msv.updateCacheHandler),
		"/update/disks": server.NewViewHandle(msv.updateDiskStatsHandler),
	}

}

func (msv *MoveServerView) GetMenu() (string, map[string]string) {
	return "Move Server", map[string]string{
		"Move Dashboard": "/",
	}
}

func New(server server.HTTPServer, moveServer *moveserver.MoveServer) (server.View, error) {
	if moveServer == nil {
		return nil, fmt.Errorf("Move server cannot be null")
	}
	return &MoveServerView{
		moveServer: moveServer,
	}, nil
}

func (msv *MoveServerView) movePostHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	log.Printf("Received move POST request", r)
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	form := r.Form
	messageType := form.Get("type")
	switch messageType {
	case "":
		err = fmt.Errorf("Submitted POST is missing type value")
	case "move":
		err = msv.moveServer.Move(form.Get("path"), form.Get("target"))
	default:
		err = fmt.Errorf("Submitted POST has unrecognized type value: %s", messageType)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (msv *MoveServerView) updateCacheHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	log.Printf("Received cache update POST request %v", r)
	msv.moveServer.UpdateCacheAsync()
	http.Redirect(w, r, "/", http.StatusFound)
}

func (msv *MoveServerView) updateDiskStatsHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	log.Printf("Received disk stats POST request %v", r)
	msv.moveServer.UpdateDiskStatsAsync()
	http.Redirect(w, r, "/", http.StatusFound)
}

func (msv *MoveServerView) pathInfoPageHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	// ---- prepare pathInfo and pathInfoHistory
	pathInfo, pathInfoHistory := msv.moveServer.GetPathInfoAndPathInfoHistory()
	pathInfoList := PathInfoSlice{}
	for _, pi := range pathInfo {
		pathInfoList = append(pathInfoList, pi)
	}
	pathInfoHistoryList := PathInfoSlice{}
	for _, pi := range pathInfoHistory {
		pathInfoHistoryList = append(pathInfoHistoryList, pi)
	}

	moveTargets := msv.moveServer.GetMoveTargets()

	mvBuffSize, mvBuffElems := msv.moveServer.GetMvBuffSizeAndElems()

	formattedCacheRefreshed := msv.moveServer.GetCacheRefreshed().Format("15:04:05 02-01-2006")

	sort.Sort(pathInfoList)
	context := struct {
		PathInfo        []moveserver.PathInfo
		PathInfoHistory []moveserver.PathInfo
		MoveTargets     []string
		Errors          []error
		CacheResfreshed string
		MvBufferSize    int
		MvBufferElems   int
		DiskStats       []moveserver.DiskStats
	}{
		PathInfo:        pathInfoList,
		PathInfoHistory: pathInfoHistoryList,
		MoveTargets:     moveTargets,
		CacheResfreshed: formattedCacheRefreshed,
		MvBufferSize:    mvBuffSize,
		MvBufferElems:   mvBuffElems,
		DiskStats:       msv.moveServer.GetDiskStats(),
	}
	s.RenderTemplate(w, r, msv.GetName(), "torrents_page", "Torrents", context)
}

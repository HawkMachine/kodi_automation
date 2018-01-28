package uploadtorrentview

import (
	"fmt"
	"net/http"

	"github.com/HawkMachine/kodi_automation/moveserver"
	"github.com/HawkMachine/kodi_automation/platform"
	"github.com/HawkMachine/kodi_automation/server"
	"github.com/HawkMachine/transmission_go_api"
)

// PathInfoSlice is sortable list of moveserver.PathInfo.
type PathInfoSlice []*moveserver.PathInfo

// Make slice of PathInfo sortable.
func (pi PathInfoSlice) Len() int {
	return len(pi)
}
func (pi PathInfoSlice) Less(i, j int) bool {
	return pi[i].Name < pi[j].Name
}
func (pi PathInfoSlice) Swap(i, j int) {
	pi[i], pi[j] = pi[j], pi[i]
}

// Implements server.View interface.
type UploadTorrentView struct {
	p  *platform.Platform
	tr *transmission_go_api.Transmission
	ms *moveserver.MoveServer
}

func (utv *UploadTorrentView) GetName() string {
	return "uploadview"
}

func (utv *UploadTorrentView) GetTemplates() map[string][]string {
	return map[string][]string{
		"upload_torrent": []string{
			"base.html",
			"upload_torrent.html",
		},
	}
}

func (utv *UploadTorrentView) GetHandlers() map[string]server.ViewHandle {
	return map[string]server.ViewHandle{
		"/upload_torrent": server.NewViewHandle(utv.uploadTorrentHandler),
	}
}

func (utv *UploadTorrentView) GetMenu() (string, map[string]string) {
	return "Move Server", map[string]string{}
}

func New(p *platform.Platform, ms *moveserver.MoveServer) (server.View, error) {
	if p == nil {
		return nil, fmt.Errorf("Move server cannot be null")
	}
	r, err := transmission_go_api.New(
		p.Config.Transmission.Address,
		p.Config.Transmission.Username,
		p.Config.Transmission.Password,
	)
	if err != nil {
		return nil, err
	}
	return &UploadTorrentView{
		p:  p,
		tr: r,
		ms: ms,
	}, nil
}

func (utv *UploadTorrentView) uploadTorrentHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	if r.Method == http.MethodGet {
		// GET -> show the form
		moveTargets := utv.ms.GetMoveTargets()
		context := struct {
			MoveTargets []string
		}{
			MoveTargets: moveTargets,
		}
		s.RenderTemplate(w, r, utv.GetName(), "upload_torrent", "Upload Torrent", context)
	} else if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		form := r.Form
		url := form.Get("url")
		if url == "" {
			err = fmt.Errorf("Empty URL submitted")
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

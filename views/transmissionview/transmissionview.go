package transmissionview

import (
	"net/http"

	"github.com/HawkMachine/kodi_automation/server"
	"github.com/HawkMachine/transmission_go_remote"
)

type TransmissionView struct {
	tr       *remote.Remote
	torrents []*remote.Torrent
}

func (tv *TransmissionView) GetName() string {
	return "transmissionview"
}

func (tv *TransmissionView) GetTemplates() map[string][]string {
	return map[string][]string{
		"transmission": []string{
			"base.html",
			"transmission_page.html",
		},
	}
}

func (tv *TransmissionView) GetHandlers() map[string]server.ViewHandle {
	return map[string]server.ViewHandle{
		"/transmission": server.NewViewHandle(tv.transmissionPage),
	}
}

func (tv *TransmissionView) GetMenu() (string, map[string]string) {
	return "Transmission", map[string]string{
		"Transmission (2)": "/transmission",
	}
}

func (tv *TransmissionView) transmissionPage(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	var err error
	tv.torrents, err = tv.tr.ListAll()
	context := struct {
		Torrents []*remote.Torrent
		Error    error
	}{
		Torrents: tv.torrents,
		Error:    err,
	}

	s.RenderTemplate(w, r, tv.GetName(), "transmission", "Transmission", context)
}

func New(address, username, password string) *TransmissionView {
	r, _ := remote.New(address, username, password, "")
	return &TransmissionView{
		tr: r,
	}
}

package wrapview

import (
	"fmt"
	"net/http"

	"github.com/HawkMachine/kodi_automation/server"
)

type WrapView struct {
	iframeLinks map[string]string
}

func (wv *WrapView) GetName() string {
	return "wrapview"
}

func (wv *WrapView) GetTemplates() map[string][]string {
	return map[string][]string{
		"wrap": []string{
			"base.html",
			"wrap_page.html",
		},
	}
}

func (wv *WrapView) GetHandlers() map[string]server.ViewHandle {
	return map[string]server.ViewHandle{
		"/wrap/": server.NewViewHandle(wv.wrapHandler),
	}
}

func (wv *WrapView) GetMenu() (string, map[string]string) {
	return "", nil
}

func (wv *WrapView) wrapHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	name := r.URL.Path[len("/wrap/"):]
	url, ok := wv.iframeLinks[name]
	if !ok {
		http.Error(w, fmt.Sprintf("Page %v not found", name), http.StatusNotFound)
		return
	}

	context := struct {
		WrapURL string
	}{
		WrapURL: url,
	}
	s.RenderTemplate(w, r, wv.GetName(), "wrap", context)
}

func New(iframeLinks map[string]string) *WrapView {
	return &WrapView{
		iframeLinks: iframeLinks,
	}
}

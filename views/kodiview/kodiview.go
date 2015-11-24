package kodiview

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/HawkMachine/kodi_automation/server"
	"github.com/HawkMachine/kodi_go_api/v6/kodi"
)

type KodiView struct {
	k *kodi.Kodi
}

func (ksv *KodiView) GetName() string {
	return "kodiview"
}

func (ksv *KodiView) GetTemplates() map[string][]string {
	return map[string][]string{
		"kodistats": []string{
			"base.html",
			"kodistats_page.html",
		},
	}
}

func (ksv *KodiView) GetHandlers() map[string]server.ViewHandle {
	return map[string]server.ViewHandle{
		"/kodi/stats":           server.NewViewHandle(ksv.kodiStatsPageHandler),
		"/kodi/stats/_getdata/": server.NewViewHandle(ksv.kodiStatsGetDataHandler),
	}
}

func (ksv *KodiView) GetMenu() (string, map[string]string) {
	return "Kodi", map[string]string{
		"Stats": "/kodi/stats",
	}
}

func (ksv *KodiView) kodiStatsPageHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	context := struct {
	}{}

	s.RenderTemplate(w, r, ksv.GetName(), "kodistats", "Kodi Stats", context)
}

func (ksv *KodiView) kodiStatsGetDataHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	result, err := ksv.k.VideoLibrary.GetEpisodes(&kodi.VideoLibraryGetEpisodesParams{
		Properties: []kodi.VideoFieldsEpisode{
			kodi.EPISODE_FIELD_SHOW_TITLE,
			kodi.EPISODE_FIELD_TITLE,
			kodi.EPISODE_FIELD_LAST_PLAYED,
		},
		Filter: kodi.VideoLibraryCreateEpisodesFilterByField(
			kodi.EPISODE_FILTER_FIELD_LAST_PLAYED,
			kodi.OPERATOR_AFTER,
			startDate.Format("2006-01-02")),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(jsonData)
}

func New(address, username, password string) *KodiView {
	return &KodiView{k: kodi.New(address+"/jsonrpc", username, password)}
}

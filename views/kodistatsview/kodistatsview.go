package kodistatsview

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/HawkMachine/kodi_automation/kodi"
	"github.com/HawkMachine/kodi_automation/server"
)

type KodiStatsView struct {
	k *kodi.Kodi
}

func (ksv *KodiStatsView) GetName() string {
	return "kodistatsview"
}

func (ksv *KodiStatsView) GetTemplates() map[string][]string {
	return map[string][]string{
		"kodistats": []string{
			"base.html",
			"kodistats_page.html",
		},
	}
}

func (ksv *KodiStatsView) GetHandlers() map[string]server.ViewHandle {
	return map[string]server.ViewHandle{
		"/kodistats":           server.NewViewHandle(ksv.kodiStatsPageHandler),
		"/kodistats/_getdata/": server.NewViewHandle(ksv.kodiStatsGetDataHandler),
	}
}

func (ksv *KodiStatsView) GetMenu() (string, map[string]string) {
	return "Kodi Stats", map[string]string{
		"Dashboard": "/kodistats",
	}
}

func (ksv *KodiStatsView) kodiStatsPageHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	context := struct {
	}{}

	s.RenderTemplate(w, r, ksv.GetName(), "kodistats", context)
}

func (ksv *KodiStatsView) kodiStatsGetDataHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	result, err := ksv.k.VideoLibrary.GetEpisodes(&kodi.GetEpisodesParams{
		Properties: []kodi.VideoFieldsEpisode{
			kodi.EPISODE_FIELD_SHOW_TITLE,
			kodi.EPISODE_FIELD_TITLE,
			kodi.EPISODE_FIELD_LAST_PLAYED,
		},
		Filter: &kodi.GetEpisodesFilter{
			ListFilterEpisodes: &kodi.ListFilterEpisodes{
				ListFilterRuleEpisodes: &kodi.ListFilterRuleEpisodes{
					Field:    kodi.EPISODE_FILTER_FIELD_LAST_PLAYED,
					Operator: kodi.OPERATOR_AFTER,
					Value:    startDate.Format("2006-01-02"),
				},
			},
		},
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

func New(address, username, password string) *KodiStatsView {
	return &KodiStatsView{k: kodi.New(address+"/jsonrpc", username, password)}
}

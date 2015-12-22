package kodiview

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HawkMachine/kodi_automation/server"
	"github.com/HawkMachine/kodi_go_api/v6/kodi"
)

func slice2map(a []string) map[string]bool {
	r := map[string]bool{}
	for _, s := range a {
		r[s] = true
	}
	return r
}

func diff(a, b map[string]bool) map[string]bool {
	r := map[string]bool{}
	for s := range a {
		if _, ok := b[s]; !ok {
			r[s] = true
		}
	}
	return r
}

type KodiView struct {
	k               *kodi.Kodi
	targets         []string
	movieExtensions map[string]bool
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
		"kodihealth": []string{
			"base.html",
			"kodihealth_page.html",
		},
		"library_movies": []string{
			"base.html",
			"library_movies_page.html",
		},
		"library_tvshows": []string{
			"base.html",
			"library_tvshows_page.html",
		},
	}
}

func (ksv *KodiView) GetHandlers() map[string]server.ViewHandle {
	return map[string]server.ViewHandle{
		"/kodi/stats":            server.NewViewHandle(ksv.kodiStatsPageHandler),
		"/kodi/stats/_getdata/":  server.NewViewHandle(ksv.kodiStatsGetDataHandler),
		"/kodi/health":           server.NewViewHandle(ksv.kodiHealthPageHandler),
		"/kodi/library/movies":   server.NewViewHandle(ksv.kodiLibraryMoviesPageHandler),
		"/kodi/library/tv_shows": server.NewViewHandle(ksv.kodiLibraryTVShowsPageHandler),
	}
}

func (ksv *KodiView) GetMenu() (string, map[string]string) {
	return "Kodi", map[string]string{
		"Stats":             "/kodi/stats",
		"Health":            "/kodi/health",
		"Library: Movies":   "/kodi/library/movies",
		"Library: TV Shows": "/kodi/library/tv_shows",
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

func directoryListing(dirname string, extensions map[string]bool) ([]string, error) {
	var res []string
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if _, ok := extensions[filepath.Ext(path)]; ok {
			res = append(res, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func filesFromVideoLibrary(k *kodi.Kodi) ([]string, error) {
	mResp, err := k.VideoLibrary.GetMovies(
		&kodi.VideoLibraryGetMoviesParams{
			Properties: []kodi.VideoFieldsMovie{
				kodi.MOVIE_FIELD_FILE,
			},
		})
	if err != nil {
		return nil, err
	}

	eResp, err := k.VideoLibrary.GetEpisodes(
		&kodi.VideoLibraryGetEpisodesParams{
			Properties: []kodi.VideoFieldsEpisode{
				kodi.EPISODE_FIELD_FILE,
			},
		})
	if err != nil {
		return nil, err
	}

	// Combine the files into one list.
	var raw []string
	for _, res := range mResp.Result.Movies {
		raw = append(raw, res.File)
	}
	for _, res := range eResp.Result.Episodes {
		raw = append(raw, res.File)
	}
	var r []string
	for _, s := range raw {
		if strings.HasPrefix(s, "stack://") {
			for _, i := range strings.Split(s[len("stack://"):], " , ") {
				r = append(r, i)
			}
		} else {
			r = append(r, s)
		}
	}

	return r, nil
}

func (ksv *KodiView) filesOnDisk() ([]string, error) {
	var paths []string
	for _, target := range ksv.targets {
		p, err := directoryListing(target, ksv.movieExtensions)
		if err != nil {
			return nil, err
		}
		paths = append(paths, p...)
	}
	return paths, nil
}

func (ksv *KodiView) kodiHealthPageHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	kodi, err := filesFromVideoLibrary(ksv.k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	disk, err := ksv.filesOnDisk()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	kodiMap := slice2map(kodi)
	diskMap := slice2map(disk)

	kodiOnly := diff(kodiMap, diskMap)
	diskOnly := diff(diskMap, kodiMap)

	context := struct {
		FilesOnDiskMissingInKodi map[string]bool
		FilesInKodiMissingOnDisk map[string]bool
	}{
		FilesOnDiskMissingInKodi: diskOnly,
		FilesInKodiMissingOnDisk: kodiOnly,
	}

	s.RenderTemplate(w, r, ksv.GetName(), "kodihealth", "Kodi Health", context)
}

func (ksv *KodiView) kodiLibraryMoviesPageHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	mResp, err := ksv.k.VideoLibrary.GetMovies(
		&kodi.VideoLibraryGetMoviesParams{
			Properties: []kodi.VideoFieldsMovie{
				kodi.MOVIE_FIELD_TITLE,
				kodi.MOVIE_FIELD_PLOT,
				kodi.MOVIE_FIELD_IMDB_NUMBER,
				kodi.MOVIE_FIELD_RATING,
				kodi.MOVIE_FIELD_DIRECTOR,
			},
			Sort: &kodi.ListSort{
				Method: kodi.SORT_METHOD_TITLE,
			},
		})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	context := struct {
		Movies []*kodi.VideoDetailsMovie
	}{
		Movies: mResp.Result.Movies,
	}

	s.RenderTemplate(w, r, ksv.GetName(), "library_movies", "Kodi Library Movies", context)
}

func (ksv *KodiView) kodiLibraryTVShowsPageHandler(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	mResp, err := ksv.k.VideoLibrary.GetTVShows(
		&kodi.VideoLibraryGetTVShowsParams{
			Properties: []kodi.VideoFieldsTVShow{
				kodi.TV_SHOW_FIELD_TITLE,
				kodi.TV_SHOW_FIELD_IMDB_NUMBER,
				kodi.TV_SHOW_FIELD_PLOT,
				kodi.TV_SHOW_FIELD_RATING,
			},
			Sort: &kodi.ListSort{
				Method: kodi.SORT_METHOD_TITLE,
			},
		})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	context := struct {
		TVShows []*kodi.VideoDetailsTVShow
	}{
		TVShows: mResp.Result.TVShows,
	}

	s.RenderTemplate(w, r, ksv.GetName(), "library_tvshows", "Kodi Library Movies", context)
}

func New(address, username, password string, targets []string) *KodiView {
	return &KodiView{
		k:       kodi.New(address+"/jsonrpc", username, password),
		targets: targets,
		movieExtensions: map[string]bool{
			".mkv":  true,
			".mp4":  true,
			".avi":  true,
			".m4v":  true,
			".ogm":  true,
			".rmvb": true,
		},
	}
}

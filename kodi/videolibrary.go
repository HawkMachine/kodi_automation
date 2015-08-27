package kodi

type VideoFieldsMovies string

const (
	MOVIE_FIELD_TITLE          VideoFieldsMovies = "title"
	MOVIE_FIELD_GENRE          VideoFieldsMovies = "genre"
	MOVIE_FIELD_YEAR           VideoFieldsMovies = "year"
	MOVIE_FIELD_RATING         VideoFieldsMovies = "rating"
	MOVIE_FIELD_DIRECTOR       VideoFieldsMovies = "director"
	MOVIE_FIELD_TRAILER        VideoFieldsMovies = "trailer"
	MOVIE_FIELD_TAGLINE        VideoFieldsMovies = "tagline"
	MOVIE_FIELD_PLOT           VideoFieldsMovies = "plot"
	MOVIE_FIELD_PLOT_OUTLINE   VideoFieldsMovies = "plotoutline"
	MOVIE_FIELD_ORIGINAL_TITLE VideoFieldsMovies = "originaltitle"
	MOVIE_FIELD_LAST_PLAYED    VideoFieldsMovies = "lastplayed"
	MOVIE_FIELD_PLAYCOUNT      VideoFieldsMovies = "playcount"
	MOVIE_FIELD_WRITER         VideoFieldsMovies = "writer"
	MOVIE_FIELD_STUDIO         VideoFieldsMovies = "studio"
	MOVIE_FIELD_MPAA           VideoFieldsMovies = "mpaa"
	MOVIE_FIELD_CAST           VideoFieldsMovies = "cast"
	MOVIE_FIELD_COUNTRY        VideoFieldsMovies = "country"
	MOVIE_FIELD_IMDB_NUMBER    VideoFieldsMovies = "imdbnumber"
	MOVIE_FIELD_RUNTIME        VideoFieldsMovies = "runtime"
	MOVIE_FIELD_SET            VideoFieldsMovies = "set"
	MOVIE_FIELD_SHOWLINK       VideoFieldsMovies = "showlink"
	MOVIE_FIELD_STREAM_DETAILS VideoFieldsMovies = "streamdetails"
	MOVIE_FIELD_TOP250         VideoFieldsMovies = "top250"
	MOVIE_FIELD_VOTES          VideoFieldsMovies = "votes"
	MOVIE_FIELD_FAN_ART        VideoFieldsMovies = "fanart"
	MOVIE_FIELD_THUMBNAIL      VideoFieldsMovies = "thumbnail"
	MOVIE_FIELD_FILE           VideoFieldsMovies = "file"
	MOVIE_FIELD_SORT_TITLE     VideoFieldsMovies = "sorttitle"
	MOVIE_FIELD_RESUME         VideoFieldsMovies = "resume"
	MOVIE_FIELD_SET_ID         VideoFieldsMovies = "setid"
	MOVIE_FIELD_DATE_ADDED     VideoFieldsMovies = "dateadded"
	MOVIE_FIELD_TAG            VideoFieldsMovies = "tag"
	MOVIE_FIELD_ART            VideoFieldsMovies = "art"
)

type ListFilterFieldsMovies string

const (
	MOVIE_FILTER_FIELD_TITLE             ListFilterFieldsMovies = "title"
	MOVIE_FILTER_FIELD_PLOT              ListFilterFieldsMovies = "plot"
	MOVIE_FILTER_FIELD_PLOT_OUTLINE      ListFilterFieldsMovies = "plotoutline"
	MOVIE_FILTER_FIELD_TAGLINE           ListFilterFieldsMovies = "tagline"
	MOVIE_FILTER_FIELD_VOTES             ListFilterFieldsMovies = "votes"
	MOVIE_FILTER_FIELD_RATING            ListFilterFieldsMovies = "rating"
	MOVIE_FILTER_FIELD_TIME              ListFilterFieldsMovies = "time"
	MOVIE_FILTER_FIELD_WRITERS           ListFilterFieldsMovies = "writers"
	MOVIE_FILTER_FIELD_PLAY_COUNT        ListFilterFieldsMovies = "playcount"
	MOVIE_FILTER_FIELD_LAST_PLAYED       ListFilterFieldsMovies = "lastplayed"
	MOVIE_FILTER_FIELD_IN_PROGRESS       ListFilterFieldsMovies = "inprogress"
	MOVIE_FILTER_FIELD_GENRE             ListFilterFieldsMovies = "genre"
	MOVIE_FILTER_FIELD_COUNTRY           ListFilterFieldsMovies = "country"
	MOVIE_FILTER_FIELD_YEAR              ListFilterFieldsMovies = "year"
	MOVIE_FILTER_FIELD_DIRECTOR          ListFilterFieldsMovies = "director"
	MOVIE_FILTER_FIELD_ACTOR             ListFilterFieldsMovies = "actor"
	MOVIE_FILTER_FIELD_MPAA_RATING       ListFilterFieldsMovies = "mpaarating"
	MOVIE_FILTER_FIELD_TOP250            ListFilterFieldsMovies = "top250"
	MOVIE_FILTER_FIELD_STUDIO            ListFilterFieldsMovies = "studio"
	MOVIE_FILTER_FIELD_HAS_TRAILER       ListFilterFieldsMovies = "hastrailer"
	MOVIE_FILTER_FIELD_FILENAME          ListFilterFieldsMovies = "filename"
	MOVIE_FILTER_FIELD_PATH              ListFilterFieldsMovies = "path"
	MOVIE_FILTER_FIELD_SET               ListFilterFieldsMovies = "set"
	MOVIE_FILTER_FIELD_TAG               ListFilterFieldsMovies = "tag"
	MOVIE_FILTER_FIELD_DATE_ADDED        ListFilterFieldsMovies = "dateadded"
	MOVIE_FILTER_FIELD_VIDEO_RESOLUTION  ListFilterFieldsMovies = "videoresolution"
	MOVIE_FILTER_FIELD_AUDIO_CHANNELS    ListFilterFieldsMovies = "audiochannels"
	MOVIE_FILTER_FIELD_VIDEO_CODEC       ListFilterFieldsMovies = "videocodec"
	MOVIE_FILTER_FIELD_AUDIO_CODEC       ListFilterFieldsMovies = "audiocodec"
	MOVIE_FILTER_FIELD_AUDIO_LANGUAGE    ListFilterFieldsMovies = "audiolanguage"
	MOVIE_FILTER_FIELD_SUBTITLE_LANGUAGE ListFilterFieldsMovies = "subtitlelanguage"
	MOVIE_FILTER_FIELD_VIDEO_ASPECT      ListFilterFieldsMovies = "videoaspect"
	MOVIE_FILTER_FIELD_PLAYLIST          ListFilterFieldsMovies = "playlist"
)

type VideoFieldsEpisode string

const (
	EPISODE_FIELD_TITLE           VideoFieldsEpisode = "title"
	EPISODE_FIELD_PLOT            VideoFieldsEpisode = "plot"
	EPISODE_FIELD_VOTES           VideoFieldsEpisode = "votes"
	EPISODE_FIELD_RATING          VideoFieldsEpisode = "rating"
	EPISODE_FIELD_WRITER          VideoFieldsEpisode = "writer"
	EPISODE_FIELD_FIRST_AIRED     VideoFieldsEpisode = "firstaired"
	EPISODE_FIELD_PLAY_COUNT      VideoFieldsEpisode = "playcount"
	EPISODE_FIELD_RUNTIME         VideoFieldsEpisode = "runtime"
	EPISODE_FIELD_DIRECTOR        VideoFieldsEpisode = "director"
	EPISODE_FIELD_PRODUCTION_CODE VideoFieldsEpisode = "productioncode"
	EPISODE_FIELD_SEASON          VideoFieldsEpisode = "season"
	EPISODE_FIELD_EPISODE         VideoFieldsEpisode = "episode"
	EPISODE_FIELD_ORIGINAL_TITLE  VideoFieldsEpisode = "originaltitle"
	EPISODE_FIELD_SHOW_TITLE      VideoFieldsEpisode = "showtitle"
	EPISODE_FIELD_CAST            VideoFieldsEpisode = "cast"
	EPISODE_FIELD_STREAM_DETAILS  VideoFieldsEpisode = "streamdetails"
	EPISODE_FIELD_LAST_PLAYED     VideoFieldsEpisode = "lastplayed"
	EPISODE_FIELD_FAN_ART         VideoFieldsEpisode = "fanart"
	EPISODE_FIELD_THUMBNAIL       VideoFieldsEpisode = "thumbnail"
	EPISODE_FIELD_FILE            VideoFieldsEpisode = "file"
	EPISODE_FIELD_RESUME          VideoFieldsEpisode = "resume"
	EPISODE_FIELD_TV_SHOW_ID      VideoFieldsEpisode = "tvshowid"
	EPISODE_FIELD_DATE_ADDED      VideoFieldsEpisode = "dateadded"
	EPISODE_FIELD_UNIQUE_ID       VideoFieldsEpisode = "uniqueid"
	EPISODE_FIELD_ART             VideoFieldsEpisode = "art"
)

type ListFilterFieldsEpisodes string

const (
	EPISODE_FILTER_FIELD_TITLE             ListFilterFieldsEpisodes = "title"
	EPISODE_FILTER_FIELD_TV_SHOW           ListFilterFieldsEpisodes = "tvshow"
	EPISODE_FILTER_FIELD_PLOT              ListFilterFieldsEpisodes = "plot"
	EPISODE_FILTER_FIELD_VOTES             ListFilterFieldsEpisodes = "votes"
	EPISODE_FILTER_FIELD_RATING            ListFilterFieldsEpisodes = "rating"
	EPISODE_FILTER_FIELD_TIME              ListFilterFieldsEpisodes = "time"
	EPISODE_FILTER_FIELD_WRITERS           ListFilterFieldsEpisodes = "writers"
	EPISODE_FILTER_FIELD_AIR_DATE          ListFilterFieldsEpisodes = "airdate"
	EPISODE_FILTER_FIELD_PLAYCOUNT         ListFilterFieldsEpisodes = "playcount"
	EPISODE_FILTER_FIELD_LAST_PLAYED       ListFilterFieldsEpisodes = "lastplayed"
	EPISODE_FILTER_FIELD_IN_PROGRESS       ListFilterFieldsEpisodes = "inprogress"
	EPISODE_FILTER_FIELD_GENRE             ListFilterFieldsEpisodes = "genre"
	EPISODE_FILTER_FIELD_YEAR              ListFilterFieldsEpisodes = "year"
	EPISODE_FILTER_FIELD_DIRECTOR          ListFilterFieldsEpisodes = "director"
	EPISODE_FILTER_FIELD_ACTOR             ListFilterFieldsEpisodes = "actor"
	EPISODE_FILTER_FIELD_EPISODE           ListFilterFieldsEpisodes = "episode"
	EPISODE_FILTER_FIELD_SEASON            ListFilterFieldsEpisodes = "season"
	EPISODE_FILTER_FIELD_FILENAME          ListFilterFieldsEpisodes = "filename"
	EPISODE_FILTER_FIELD_PATH              ListFilterFieldsEpisodes = "path"
	EPISODE_FILTER_FIELD_STUDIO            ListFilterFieldsEpisodes = "studio"
	EPISODE_FILTER_FIELD_MPAA_RATING       ListFilterFieldsEpisodes = "mpaarating"
	EPISODE_FILTER_FIELD_DATE_ADDED        ListFilterFieldsEpisodes = "dateadded"
	EPISODE_FILTER_FIELD_VIDEO_RESOLUTION  ListFilterFieldsEpisodes = "videoresolution"
	EPISODE_FILTER_FIELD_AUDIO_CHANNELS    ListFilterFieldsEpisodes = "audiochannels"
	EPISODE_FILTER_FIELD_VIDEO_CODEC       ListFilterFieldsEpisodes = "videocodec"
	EPISODE_FILTER_FIELD_AUDIO_CODEC       ListFilterFieldsEpisodes = "audiocodec"
	EPISODE_FILTER_FIELD_AUDIO_LANGUAGE    ListFilterFieldsEpisodes = "audiolanguage"
	EPISODE_FILTER_FIELD_SUBTITLE_LANGUAGE ListFilterFieldsEpisodes = "subtitlelanguage"
	EPISODE_FILTER_FIELD_VIDEO_ASPECT      ListFilterFieldsEpisodes = "videoaspect"
	EPISODE_FILTER_FIELD_PLAYLIST          ListFilterFieldsEpisodes = "playlist"
)

type VideoFieldsSeason string

const (
	SEASON_FIELD_SEASON           VideoFieldsSeason = "season"
	SEASON_FIELD_SHOW_TITLE       VideoFieldsSeason = "showtitle"
	SEASON_FIELD_PLAY_COUNT       VideoFieldsSeason = "playcount"
	SEASON_FIELD_EPISODE          VideoFieldsSeason = "episode"
	SEASON_FIELD_FAN_ART          VideoFieldsSeason = "fanart"
	SEASON_FIELD_THUMBNAIL        VideoFieldsSeason = "thumbnail"
	SEASON_FIELD_TV_SHOW_ID       VideoFieldsSeason = "tvshowid"
	SEASON_FIELD_WATCHED_EPISODES VideoFieldsSeason = "watchedepisodes"
	SEASON_FIELD_ART              VideoFieldsSeason = "art"
)

type VideoFieldsTvShow string

const (
	TV_SHOW_FIELD_TITLE            VideoFieldsTvShow = "title"
	TV_SHOW_FIELD_GENRE            VideoFieldsTvShow = "genre"
	TV_SHOW_FIELD_YEAR             VideoFieldsTvShow = "year"
	TV_SHOW_FIELD_RATING           VideoFieldsTvShow = "rating"
	TV_SHOW_FIELD_PLOT             VideoFieldsTvShow = "plot"
	TV_SHOW_FIELD_STUDIO           VideoFieldsTvShow = "studio"
	TV_SHOW_FIELD_MPAA             VideoFieldsTvShow = "mpaa"
	TV_SHOW_FIELD_CAST             VideoFieldsTvShow = "cast"
	TV_SHOW_FIELD_PLAY_COUTN       VideoFieldsTvShow = "playcount"
	TV_SHOW_FIELD_EPISODE          VideoFieldsTvShow = "episode"
	TV_SHOW_FIELD_IMDB_NUMBER      VideoFieldsTvShow = "imdbnumber"
	TV_SHOW_FIELD_PREMERED_AT      VideoFieldsTvShow = "premiered"
	TV_SHOW_FIELD_VOTES            VideoFieldsTvShow = "votes"
	TV_SHOW_FIELD_LAST_PLAYED      VideoFieldsTvShow = "lastplayed"
	TV_SHOW_FIELD_FAN_ART          VideoFieldsTvShow = "fanart"
	TV_SHOW_FIELD_THUMBNAIL        VideoFieldsTvShow = "thumbnail"
	TV_SHOW_FIELD_FILE             VideoFieldsTvShow = "file"
	TV_SHOW_FIELD_ORIGINAL_TITLE   VideoFieldsTvShow = "originaltitle"
	TV_SHOW_FIELD_SORT_TITLE       VideoFieldsTvShow = "sorttitle"
	TV_SHOW_FIELD_EPISODE_GUIDE    VideoFieldsTvShow = "episodeguide"
	TV_SHOW_FIELD_SEASON           VideoFieldsTvShow = "season"
	TV_SHOW_FIELD_WATCHED_EPISODES VideoFieldsTvShow = "watchedepisodes"
	TV_SHOW_FIELD_DATE_ADDED       VideoFieldsTvShow = "dateadded"
	TV_SHOW_FIELD_TAG              VideoFieldsTvShow = "tag"
	TV_SHOW_FIELD_ART              VideoFieldsTvShow = "art"
)

type ListFilterFieldsTVShows string

const (
	TV_SHOW_FILTER_FIELD_TITLE        ListFilterFieldsTVShows = "title"
	TV_SHOW_FILTER_FIELD_PLOT         ListFilterFieldsTVShows = "plot"
	TV_SHOW_FILTER_FIELD_STATUS       ListFilterFieldsTVShows = "status"
	TV_SHOW_FILTER_FIELD_VOTES        ListFilterFieldsTVShows = "votes"
	TV_SHOW_FILTER_FIELD_RATING       ListFilterFieldsTVShows = "rating"
	TV_SHOW_FILTER_FIELD_YEAR         ListFilterFieldsTVShows = "year"
	TV_SHOW_FILTER_FIELD_GENRE        ListFilterFieldsTVShows = "genre"
	TV_SHOW_FILTER_FIELD_DIRECTOR     ListFilterFieldsTVShows = "director"
	TV_SHOW_FILTER_FIELD_ACTOR        ListFilterFieldsTVShows = "actor"
	TV_SHOW_FILTER_FIELD_NUM_EPISODES ListFilterFieldsTVShows = "numepisodes"
	TV_SHOW_FILTER_FIELD_NUM_WATCHED  ListFilterFieldsTVShows = "numwatched"
	TV_SHOW_FILTER_FIELD_PLAY_COUNT   ListFilterFieldsTVShows = "playcount"
	TV_SHOW_FILTER_FIELD_PATH         ListFilterFieldsTVShows = "path"
	TV_SHOW_FILTER_FIELD_STUDIO       ListFilterFieldsTVShows = "studio"
	TV_SHOW_FILTER_FIELD_MPAA_RATING  ListFilterFieldsTVShows = "mpaarating"
	TV_SHOW_FILTER_FIELD_DATE_ADDED   ListFilterFieldsTVShows = "dateadded"
	TV_SHOW_FILTER_FIELD_LAST_PLAYED  ListFilterFieldsTVShows = "lastplayed"
	TV_SHOW_FILTER_FIELD_IN_PROGRESS  ListFilterFieldsTVShows = "inprogress"
	TV_SHOW_FILTER_FIELD_PLAYLIST     ListFilterFieldsTVShows = "playlist"
)

type VideoCastElement struct {
	Thumbnail string `json:"thumbnail"`
	Name      string `json:"name"`
	Role      string `json:"role"`
}

type VideoDetailsBase struct {
	MediaDetailsBase
	PlayCount int `json:"playcount"`
}

type VideoDetailsMedia struct {
	VideoDetailsBase
	Title string `json:"title"`
}

type VideoDetailsItem struct {
	VideoDetailsMedia
	DateAdded  string `json:"dateadded,omitempty"`
	File       string `json:"file,omitempty"`
	LastPlayed string `json:"lastplayed,omitempty"`
	Plot       string `json:"plot,omitempty"`
}

type VideoDetailsFile struct {
	VideoDetailsItem
}

type VideoDetailsMovie struct {
	VideoDetailsFile
	PlotOutline   string              `json:"plotoutline,omitempty"`
	SortTitle     string              `json:"sorttitle,omitempty"`
	MovieId       LibraryId           `json:"movieid,omitempty"`
	Cast          []*VideoCastElement `json:"cast,omitempty"`
	Votes         string              `json:"votes,omitempty"`
	ShowLink      []string            `json:"showlink,omitempty"`
	Top250        int                 `json:"top250,omitempty"`
	Trailer       string              `json:"trailer,omitempty"`
	Year          int                 `json:"year,omitempty"`
	Country       []string            `json:"country,omitempty"`
	Studio        []string            `json:"studio,omitempty"`
	Set           string              `json:"set,omitempty"`
	Genre         []string            `json:"genre,omitempty"`
	MPAA          string              `json:"mpaa,omitempty"`
	SetId         LibraryId           `json:"setid,omitempty"`
	Rating        float32             `json:"rating,omitempty"`
	Tag           string              `json:"tag,omitempty"`
	TagLine       string              `json:"tagline,omitempty"`
	Writer        []string            `json:"writer,omitempty"`
	OriginalTitle string              `json:"originaltitle,omitempty"`
	IMDBNumber    string              `json:"imdbnumber,omitempty"`
}

type VideoDetailsEpisode struct {
	VideoDetailsFile
	Cast           []*VideoCastElement `json:"cast,omitempty"`
	ProductionCode string              `json:"productioncode,omitempty"`
	Rating         float32             `json:"rating,omitempty"`
	Votes          string              `json:"votes,omitempty"`
	Episode        int                 `json:"episode,omitempty"`
	ShowTitle      string              `json:"showtitle,omitempty"`
	EpisodeId      LibraryId           `json:"episodeid,omitempty"`
	TVShowId       LibraryId           `json:"tvshowid,omitempty"`
	Season         int                 `json:"season,omitempty"`
	FirstAired     string              `json:"firstaired,omitempty"`
	UniqueId       interface{}         `json:"uniqueid,omitempty"`
	Writer         []string            `json:"writer,omitempty"`
	OriginalTitle  string              `json:"originaltitle,omitempty"`
}

type VideoDetailsSeason struct {
	VideoDetailsBase
	ShowTitle       string    `json:"showtitle,omitempty"`
	WatchedEpisodes int       `json:"watchedepisodes,omitempty"`
	TVShowId        LibraryId `json:"tvshowid,omitempty"`
	Episode         int       `json:"episode,omitempty"`
	Season          int       `json:"season,omitempty"`
}

type VideoDetailsTVShow struct {
	VideoDetailsItem
	SortTitle       string              `json:"sorttitle"`
	MPAA            string              `json:"mpaa"`
	Premiered       string              `json:"premiered"`
	Year            int                 `json:"year"`
	Episode         int                 `json:"episode"`
	WatchedEpisodes int                 `json:"watchedepisodes"`
	Votes           string              `json:"votes"`
	Rating          float32             `json:"rating"`
	TVShowId        LibraryId           `json:"tvshowid"`
	Studio          []string            `json:"studio"`
	Season          int                 `json:"season"`
	Genre           []string            `json:"genre"`
	Cast            []*VideoCastElement `json:"cast"`
	EpisodeGuide    string              `json:"episodeguide"`
	Tag             []string            `json:"tag"`
	OriginalTitle   string              `json:"originaltitle"`
	IMDBNumber      string              `json:"imdbnumber"`
}

// Implementation of VideoLibrary RPCs.
type VideoLibrary struct {
	k *Kodi
}

func (vl *VideoLibrary) doRPC(rpcName string, params interface{}, response interface{}) error {
	req := &request{
		Jsonrpc: "2.0",
		Method:  "VideoLibrary." + rpcName,
		Id:      "1",
		Params:  params,
	}
	return vl.k.doRPC(req, response)
}

// VideoLibrary.GetMovies

type ListFilterRuleMovies struct {
	Field    string             `json:"field,omitempty"`
	Operator ListFilterOperator `json:"operator"`
	Value    string             `json:"value"`
}

type ListFilterMovies struct {
	// one of:
	*ListFilterRuleMovies
	And []*ListFilterMovies `json:"and,omitempty"`
	Or  []*ListFilterMovies `json:"or,omitempty"`
}

type GetMoviesFilter struct {
	// one of:
	*ListFilterMovies
	GenreId  *LibraryId `json:"genreid,omitempty"`
	Genre    *string    `json:"genre,omitempty"`
	Year     *int       `json:"year,omitempty"`
	Actor    *string    `json:"actor,omitempty"`
	Director *string    `json:"director,omitempty"`
	Studio   *string    `json:"studio,omitempty"`
	Country  *string    `json:"country,omitempty"`
	SetId    *LibraryId `json:"setid,omitempty"`
	Set      *string    `json:"set,omitempty"`
	Tag      *string    `json:"tag,omitempty"`
}

type GetMoviesParams struct {
	Properties []VideoFieldsMovies `json:"properties"`
	Limits     *ListLimits         `json:"limits,omitempty"`
	Sort       *ListSort           `json:"sort,omitempty"`
	Filter     *GetMoviesFilter    `json:"filter,omitempty"`
}

type GetMoviesResult struct {
	Limits ListLimitsReturned   `json:"limits,omitempty"`
	Movies []*VideoDetailsMovie `json:"movies,omitempty"`
}

type GetMoviesResponse struct {
	ResponseBase
	Result *GetMoviesResult `json:"result,omitempty"`
}

func (vl *VideoLibrary) GetMovies(params *GetMoviesParams) (*GetMoviesResponse, error) {
	response := &GetMoviesResponse{}
	err := vl.doRPC("GetMovies", params, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// VideoLibrary.GetMovieDetails

type GetMovieDetailsParams struct {
	MovieId    LibraryId           `json:"movieid"`
	Properties []VideoFieldsMovies `json:"properties,omitempty"`
}

type GetMovieDetailsResult struct {
	MovieDetails *VideoDetailsMovie `json:"moviedetails,omitempty"`
}

type GetMovieDetailsResponse struct {
	ResponseBase
	Result *GetMovieDetailsResult
}

func (vl *VideoLibrary) GetMovieDetails(params *GetMovieDetailsParams) (*GetMovieDetailsResponse, error) {
	response := &GetMovieDetailsResponse{}
	err := vl.doRPC("GetMovieDetailsResponse", params, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// VideoLibrary.GetEpisodes

type ListFilterRuleEpisodes struct {
	Field    ListFilterFieldsEpisodes `json:"field,omitempty"`
	Operator ListFilterOperator       `json:"operator"`
	Value    string                   `json:"value"`
}

type ListFilterEpisodes struct {
	// one of:
	*ListFilterRuleEpisodes
	And []*ListFilterMovies `json:"and,omitempty"`
	Or  []*ListFilterMovies `json:"or,omitempty"`
}

type GetEpisodesFilter struct {
	// one of:
	*ListFilterEpisodes
	GenreId  *LibraryId `json:"genreid,omitempty"`
	Genre    *string    `json:"genre,omitempty"`
	Year     *int       `json:"year,omitempty"`
	Actor    *string    `json:"actor,omitempty"`
	Director *string    `json:"director,omitempty"`
}

type GetEpisodesParams struct {
	TVShowId   *LibraryId           `json:"tvshowid,omitempty"`
	Season     *int                 `json:"season,omitempty"`
	Properties []VideoFieldsEpisode `json:"properties,omitempty"`
	Limits     *ListLimits          `json:"limits,omitempty"`
	Sort       *ListSort            `json:"sort,omitempty"`
	Filter     *GetEpisodesFilter   `json:"filter,omitempty"`
}

type GetEpisodesResult struct {
	Episodes []*VideoDetailsEpisode `json:"episodes,omitempty"`
	Limits   ListLimitsReturned     `json:"limits"`
}

type GetEpisodesResponse struct {
	ResponseBase
	Result *GetEpisodesResult `json:"result,omitempty"`
}

func (vl *VideoLibrary) GetEpisodes(params *GetEpisodesParams) (*GetEpisodesResponse, error) {
	response := &GetEpisodesResponse{}
	err := vl.doRPC("GetEpisodes", params, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// VideoLibrary.GetTVShows

type GetTVShowsParams struct {
	Properties []VideoFieldsTvShow `json:"properties,omitempty"`
	Limits     *ListLimits         `json:"limits,omitempty"`
	Sort       *ListSort           `json:"sort,omitempty"`
	Filter     *Filter             `json:"filter,omitempty"`
}

type GetTVShowsResult struct {
	Limits  ListLimitsReturned    `json:"limits,omitempty"`
	TVShows []*VideoDetailsTVShow `json:"tvshows,omitempty"`
}

type GetTVShowsResponse struct {
	ResponseBase
	Result *GetTVShowsResult `json:"result,omitempty"`
}

func (vl *VideoLibrary) GetTVShows(params *GetTVShowsParams) (*GetTVShowsResponse, error) {
	response := &GetTVShowsResponse{}
	err := vl.doRPC("GetTVShows", params, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// VideoLibrary.GetTVShowDetails

type GetTVShowDetailsParams struct {
	TVShowId   LibraryId           `json:"tvshowid"`
	Properties []VideoFieldsTvShow `json:"properties,omitempty"`
}

type GetTVShowDetailsResult struct {
	TVShowDetails *VideoDetailsTVShow `json:"tvshowdetails,omitempty"`
}

type GetTVShowDetailsResponse struct {
	ResponseBase
	Result *GetTVShowDetailsResult `json:"result,omitempty"`
}

func (vl *VideoLibrary) GetTVShowDetails(params *GetTVShowDetailsParams) (*GetTVShowDetailsResponse, error) {
	response := &GetTVShowDetailsResponse{}
	err := vl.doRPC("GetTVShowDetails", params, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// VideoLibrary.GetSeasons

type GetSeasonsParams struct {
	TVShowId   LibraryId           `json:"tvshowid"`
	Properties []VideoFieldsSeason `json:"properties,omitempty"`
	Limits     *ListLimits         `json:"limits,omitempty"`
	Sort       *ListSort           `json:"sort,omitempty"`
}

type GetSeasonsResult struct {
	Limits  ListLimitsReturned    `json:"limits,omitempty"`
	Seasons []*VideoDetailsSeason `json:"seasons,omitempty"`
}

type GetSeasonsResponse struct {
	ResponseBase
	Result *GetSeasonsResult `json:"result,omitempty"`
}

func (vl *VideoLibrary) GetSeasons(params *GetSeasonsParams) (*GetSeasonsResponse, error) {
	response := &GetSeasonsResponse{}
	err := vl.doRPC("GetSeasons", params, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

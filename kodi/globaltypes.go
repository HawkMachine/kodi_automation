package kodi

type ListFilterOperator string

const (
	OPERATOR_CONTAINS         ListFilterOperator = "contains"
	OPERATOR_DOES_NOT_CONTAIN ListFilterOperator = "doesnotcontain"
	OPERATOR_IS               ListFilterOperator = "is"
	OPERATOR_IS_NOT           ListFilterOperator = "isnot"
	OPERATOR_STARTS_WITH      ListFilterOperator = "startswith"
	OPERATOR_ENDS_WITH        ListFilterOperator = "endswith"
	OPERATOR_GREATER_THAN     ListFilterOperator = "greaterthan"
	OPERATOR_LESS_THAN        ListFilterOperator = "lessthan"
	OPERATOR_AFTER            ListFilterOperator = "after"
	OPERATOR_BEFORE           ListFilterOperator = "before"
	OPERATOR_IN_THE_LAST      ListFilterOperator = "inthelast"
	OPERATOR_NOT_INT_THE_LAST ListFilterOperator = "notinthelast"
	OPERATOR_TRUE             ListFilterOperator = "true"
	OPERATOR_FALSE            ListFilterOperator = "false"
	OPERATOR_BETWEEN          ListFilterOperator = "between"
)

type SortOrder string

const (
	SORT_ORDER_ASCENDING  SortOrder = "ascending"
	SORT_ORDER_DESCENDING SortOrder = "descending"
)

type SortMethod string

const (
	SORT_METHOD_NONE             SortMethod = "none"
	SORT_METHOD_LABEL            SortMethod = "label"
	SORT_METHOD_DATE             SortMethod = "date"
	SORT_METHOD_SIZE             SortMethod = "size"
	SORT_METHOD_FILE             SortMethod = "file"
	SORT_METHOD_PATH             SortMethod = "path"
	SORT_METHOD_DRIVE_TYPE       SortMethod = "drivetype"
	SORT_METHOD_TITLE            SortMethod = "title"
	SORT_METHOD_TRACK            SortMethod = "track"
	SORT_METHOD_TIME             SortMethod = "time"
	SORT_METHOD_ARTIST           SortMethod = "artist"
	SORT_METHOD_ALBUM            SortMethod = "album"
	SORT_METHOD_ALBUM_TYPE       SortMethod = "albumtype"
	SORT_METHOD_GENRE            SortMethod = "genre"
	SORT_METHOD_COUNTRY          SortMethod = "country"
	SORT_METHOD_YEAR             SortMethod = "year"
	SORT_METHOD_RATING           SortMethod = "rating"
	SORT_METHOD_VOTES            SortMethod = "votes"
	SORT_METHOD_TOP250           SortMethod = "top250"
	SORT_METHOD_PROGRAM_COUNT    SortMethod = "programcount"
	SORT_METHOD_PLAYLIST         SortMethod = "playlist"
	SORT_METHOD_EPISODE          SortMethod = "episode"
	SORT_METHOD_SEASON           SortMethod = "season"
	SORT_METHOD_TOTAL_EPISODES   SortMethod = "totalepisode"
	SORT_METHOD_WATCHED_EPISODES SortMethod = "watchedepisodes"
	SORT_METHOD_TV_SHOW_STATUS   SortMethod = "tvshowstatus"
	SORT_METHOD_TV_SHOW_TITLE    SortMethod = "tvshowtitle"
	SORT_METHOD_SORT_TITLE       SortMethod = "sorttitle"
	SORT_METHOD_PRODUCTION_CODE  SortMethod = "productioncode"
	SORT_METHOD_MPAA             SortMethod = "mpaa"
	SORT_METHOD_STUDIO           SortMethod = "studio"
	SORT_METHOD_DATE_ADDED       SortMethod = "dataadded"
	SORT_METHOD_LAST_PLAYED      SortMethod = "lastplayed"
	SORT_METHOD_PLAY_COUNT       SortMethod = "playcount"
	SORT_METHOD_LISTENERS        SortMethod = "listeners"
	SORT_METHOD_BIT_RATE         SortMethod = "bitrate"
	SORT_METHOD_RANDOM           SortMethod = "random"
)

type FileType string

const (
	FILE_TYPE_FILE FileType = "file"
	FILE_TYPE_DIR  FileType = "directory"
)

type LibraryId int

type ListSort struct {
	Order         string `json:"order"`
	Ignorearticle bool   `json:"ignorearticle"`
	Method        string `json:"method"`
}

type Filter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type ListLimits struct {
	Start int `json:"start,omitempty"`
	End   int `json:"end,omitempty"`
}

type ListLimitsReturned struct {
	Total int `json:"total"`
	Start int `json:"start,omitempty"`
	End   int `json:"end,omitempty"`
}

type ItemDetailsBase struct {
	Label string `json:"label"`
}

type MediaDetailsBase struct {
	ItemDetailsBase
	Fanart    string `json:"fanart,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

func FilterString(a string) *string {
	return &a
}

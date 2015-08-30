package builder

import (
	"fmt"
	"sort"

	"github.com/HawkMachine/kodi_go_api/v6/kodi"
)

type EpisodeBy func(e1, e2 *kodi.VideoDetailsEpisode) bool

func (by EpisodeBy) Sort(episodes []*kodi.VideoDetailsEpisode) {
	rs := episodeSorter{
		episodes: episodes,
		by:       by,
	}
	sort.Sort(rs)
}

type episodeSorter struct {
	episodes []*kodi.VideoDetailsEpisode
	by       EpisodeBy
}

func (es episodeSorter) Len() int {
	return len(es.episodes)
}

func (es episodeSorter) Swap(i, j int) {
	es.episodes[i], es.episodes[j] = es.episodes[j], es.episodes[i]
}

func (es episodeSorter) Less(i, j int) bool {
	return es.by(es.episodes[i], es.episodes[j])
}

func EpisodePositionLess(e1, e2 *kodi.VideoDetailsEpisode) bool {
	if e1.Season == e2.Season {
		return e1.Episode < e2.Episode
	}
	return e1.Season < e2.Season
}

// GetUnwatchedPlaylist returns a list (of at most num length) of episode
// details built from tv shows passed as arguments. Episodes are chosen basing
// on play count. Episodes watched the fewest time have priority.
func GetUnwatchedPlaylist(k *kodi.Kodi, tvShowsNames []string, num int) ([]*kodi.VideoDetailsEpisode, error) {
	// dedup show names
	dedupedTvShowNames := map[string]bool{}
	for _, tvShowName := range tvShowsNames {
		dedupedTvShowNames[tvShowName] = true
	}

	playCountAndPosition := func(e1, e2 *kodi.VideoDetailsEpisode) bool {
		// Sort by PlayCount first, then by position of episode in the show.
		if e1.PlayCount == e2.PlayCount {
			return EpisodePositionLess(e1, e2)
		}
		return e1.PlayCount < e2.PlayCount
	}

	// Find Episodes by TV Show
	episodesByShow := map[string][]*kodi.VideoDetailsEpisode{}
	for tvShowName := range dedupedTvShowNames {
		resp, err := k.VideoLibrary.GetEpisodes(
			&kodi.VideoLibraryGetEpisodesParams{
				Properties: []kodi.VideoFieldsEpisode{
					kodi.EPISODE_FIELD_EPISODE,
					kodi.EPISODE_FIELD_SEASON,
					kodi.EPISODE_FIELD_TITLE,
					kodi.EPISODE_FIELD_LAST_PLAYED,
					kodi.EPISODE_FIELD_PLAY_COUNT,
					kodi.EPISODE_FIELD_SHOW_TITLE,
					kodi.EPISODE_FIELD_RUNTIME,
				},
				Filter: kodi.VideoLibraryCreateEpisodesFilterByField(
					kodi.EPISODE_FILTER_FIELD_TV_SHOW,
					kodi.OPERATOR_IS,
					tvShowName),
			})
		if err != nil {
			return nil, err
		}
		if resp.Error != nil || resp.Result == nil {
			return nil, fmt.Errorf("Kodi returned error or nil result: %v, %v", resp.Error, resp.Result)
		}

		// Sort episodes by play count, then by position in the show.
		EpisodeBy(playCountAndPosition).Sort(resp.Result.Episodes)
		episodesByShow[tvShowName] = resp.Result.Episodes
	}

	// Build the result list.
	var result []*kodi.VideoDetailsEpisode
	idx := 0
	for {
		if len(result) == num {
			break
		}
		epAdded := false
		for tvShowName := range dedupedTvShowNames {
			if len(result) == num {
				break
			}
			episodes := episodesByShow[tvShowName]
			if idx >= len(episodes) {
				continue
			}
			result = append(result, episodes[idx])
			epAdded = true
		}
		if !epAdded {
			break
		}
		idx++
	}

	// Sort the result so that episodes from the same show are together.
	showAndPosition := func(e1, e2 *kodi.VideoDetailsEpisode) bool {
		// Sort by PlayCount first, then by position of episode in the show.
		if e1.ShowTitle == e2.ShowTitle {
			return EpisodePositionLess(e1, e2)
		}
		return e1.ShowTitle < e2.ShowTitle
	}
	EpisodeBy(showAndPosition).Sort(result)
	return result, nil
}

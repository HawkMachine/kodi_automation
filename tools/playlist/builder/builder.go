package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/HawkMachine/kodi_go_api/v6/kodi"

	playlist_builder "github.com/HawkMachine/kodi_automation/kodi/playlist/builder"
)

var (
	username = flag.String("u", "", "Kodi username")
	password = flag.String("p", "", "Kodi password")
	address  = flag.String("a", "http://192.168.0.200:9080", "Kodi address")

	size  = flag.Int("s", 5, "Number of episodes.")
	shows = flag.String("shows", "", "Comma-separated list of tv shows names.")

	dryRun = flag.Bool("dry_run", true, "Dry run mode - only print the playlist")
)

func main() {
	flag.Parse()
	if *address == "" || *username == "" || *password == "" {
		log.Fatalf("Address, user or password is missing: %s, %s, %s", *address, *username, *password)
	}

	k := kodi.New(*address+"/jsonrpc", *username, *password)

	tvShowNames := strings.Split(*shows, ",")
	episodes, err := playlist_builder.GetUnwatchedPlaylist(k, tvShowNames, *size)
	if err != nil {
		log.Fatal(err)
	}
	for _, ep := range episodes {
		fmt.Printf("%20s : S%2dE%2d : %q\n", ep.ShowTitle, ep.Season, ep.Episode, ep.Title)
	}

	if *dryRun {
		return
	}

	fmt.Println("==== Submitting to kodi ===")

	for _, ep := range episodes {
		resp, err := k.Playlist.Add(&kodi.PlaylistAddParams{
			PlaylistId: kodi.PLAYLIST_VIDEO_ID,
			Item: kodi.PlaylistItem{
				EpisodeId: &ep.EpisodeId,
			},
		})
		if err != nil {
			log.Fatal(err)
			return
		}
		if resp.Error != nil {
			log.Fatal(resp.Error)
		}
		fmt.Printf("%20s : S%2dE%2d : %q : %s\n", ep.ShowTitle, ep.Season, ep.Episode, ep.Title, *resp.Result)
	}
}

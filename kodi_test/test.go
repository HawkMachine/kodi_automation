package main

import (
	//"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/HawkMachine/kodi_automation/kodi"
)

var (
	username = flag.String("u", "", "Kodi username")
	password = flag.String("p", "", "Kodi password")
	address  = flag.String("a", "http://192.168.0.200:9080", "Kodi address")
)

func main() {
	flag.Parse()
	if *address == "" || *username == "" || *password == "" {
		log.Fatalf("Address, user or password is missing: %s, %s, %s", *address, *username, *password)
	}

	k := kodi.New(*address+"/jsonrpc", *username, *password)

	epResult, err := k.VideoLibrary.GetEpisodes(&kodi.GetEpisodesParams{
		Properties: []kodi.VideoFieldsEpisode{
			kodi.EPISODE_FIELD_SHOW_TITLE,
			kodi.EPISODE_FIELD_TITLE,
			kodi.EPISODE_FIELD_LAST_PLAYED,
		},
		Filter: &kodi.GetEpisodesFilter{
			ListFilterEpisodes: &kodi.ListFilterEpisodes{
				ListFilterRuleEpisodes: &kodi.ListFilterRuleEpisodes{
					Field:    kodi.EPISODE_FILTER_FIELD_LASTPLAYED,
					Operator: kodi.OPERATOR_AFTER,
					Value:    "2015-08-01",
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	byDate := map[time.Time]int{}
	for _, edetails := range epResult.Result.Episodes {
		t, err := time.Parse("2006-01-02 15:04:05", edetails.LastPlayed)
		if err != nil {
			log.Fatal(err)
		}
		year, month, day := t.Date()
		d := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		v, ok := byDate[d]
		if ok {
			byDate[d] = v + 1
		} else {
			byDate[d] = v + 1
		}
	}
	for s, v := range byDate {
		fmt.Printf("%30s : %d\n", s, v)
	}
}

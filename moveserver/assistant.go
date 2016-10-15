package moveserver

import (
	"fmt"
	"time"

	tr "github.com/HawkMachine/transmission_go_api"
)

func filterPathInfo(a map[string]*PathInfo, f func(*PathInfo) bool) []*PathInfo {
	var r []*PathInfo
	for _, pi := range a {
		if f(pi) {
			r = append(r, pi)
		}
	}
	return r
}

type Assistant struct {
	msv                   *MoveServer
	sleep                 time.Duration
	pause                 bool
	dryRun                bool
	moveTarget            string
	maxConcurrentTorrents int
}

func (a *Assistant) Log(tp, msg string) {
	a.msv.Log(fmt.Sprintf("Assistant.%s", tp), msg)
}

func (a *Assistant) run() {
	if a.sleep < 5*time.Minute {
		a.sleep = 5 * time.Minute
	}
	a.pause = false
	for !a.pause {
		err := a.assist()
		if err != nil {
			a.Log("assist", fmt.Sprintf("Failed with error: %s", err))
		}
		time.Sleep(a.sleep)
	}
}

func (a *Assistant) assist() error {
	a.msv.lock.Lock()
	defer a.msv.lock.Unlock()

	// Select candidates for moving.
	// Very, very simple logic. Check if there are any torrents, allowed to move
	// without move error. These must have torrent info and be paused (I rely
	// here on a cron job that automatically pauses finished torrents.
	pis := a.msv.pathInfo
	toMove := filterPathInfo(pis, func(pi *PathInfo) bool {
		return pi.AllowMove && !pi.MoveInfo.Moving && pi.MoveInfo.LastError != nil && pi.Torrent != nil && pi.Torrent.DoneDate != 0 && pi.Torrent.Status == tr.TR_STATUS_PAUSED
	})
	todo := filterPathInfo(pis, func(pi *PathInfo) bool {
		return pi.Torrent != nil && pi.Torrent.Status == tr.TR_STATUS_PAUSED && pi.Torrent.DoneDate == 0
	})
	downloading := filterPathInfo(pis, func(pi *PathInfo) bool {
		return pi.Torrent != nil && pi.Torrent.Status != tr.TR_STATUS_PAUSED
	})

	for _, pi := range toMove {
		a.Log("assist", fmt.Sprintf("Would like to move %s to %s", pi.Name, a.moveTarget))
	}

	if len(downloading) > 0 {
		if len(downloading) > a.maxConcurrentTorrents {
			a.Log("assist", fmt.Sprintf("Downloading %d, max concurrent downloads = %d, cannot download more", len(downloading), a.maxConcurrentTorrents))
		} else {
			a.Log("assist", fmt.Sprintf("Downloading %d, max concurrent downloads = %d, can enable %d out of %d",
				len(downloading), a.maxConcurrentTorrents, a.maxConcurrentTorrents-len(downloading), len(todo)))
		}
	} else {
		a.Log("assist", "There are no candidates to download")
	}

	return nil
}

package moveserver

import (
	"fmt"
	"sync"
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
	msv                      *MoveServer
	sleep                    time.Duration
	enabled                  bool
	runStarted               bool
	dryRun                   bool
	moveTarget               string
	maxConcurrentDownloading int
	maxConcurrentMoving      int

	lock sync.Mutex
}

func (a *Assistant) Log(tp, msg string) {
	a.msv.Log(fmt.Sprintf("Assistant.%s", tp), msg)
}

func (a *Assistant) Enable() {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.enabled {
		return
	}
	if !a.runStarted {
		go a.run()
		a.runStarted = true
	}
	a.enabled = true
}

func (a *Assistant) Disable() {
	a.lock.Lock()
	defer a.lock.Unlock()

	if !a.enabled {
		return
	}
	a.enabled = false
}

func (a *Assistant) IsEnabled() bool {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.enabled
}

func (a *Assistant) run() {
	if a.sleep < 1*time.Minute {
		a.sleep = 1 * time.Minute
	}
	// a.sleep = 15 * time.Second
	for !a.enabled {
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
	// here on a cron job that automatically pause finished torrents.
	pis := a.msv.pathInfo
	toMove := filterPathInfo(pis, func(pi *PathInfo) bool {
		return pi.AllowMove && !pi.MoveInfo.Moving && pi.Path != "" && pi.MoveInfo.LastError == nil && pi.Torrent != nil && pi.Torrent.DoneDate != 0 && pi.Torrent.PercentDone == 1.0 && pi.Torrent.Status == tr.TR_STATUS_PAUSED
	})
	todo := filterPathInfo(pis, func(pi *PathInfo) bool {
		return pi.Torrent != nil && pi.Torrent.Status == tr.TR_STATUS_PAUSED && pi.Torrent.DoneDate == 0 && pi.Torrent.PercentDone != 1.0
	})

	downloading := filterPathInfo(pis, func(pi *PathInfo) bool {
		return pi.Torrent != nil && pi.Torrent.Status != tr.TR_STATUS_PAUSED
	})
	moving := filterPathInfo(pis, func(pi *PathInfo) bool {
		return pi.MoveInfo.Moving
	})

	// *** Request to move some torrents.
	toMoveSelected := []*PathInfo{}
	for _, pi := range toMove {
		if len(toMoveSelected)+len(moving) >= a.maxConcurrentMoving {
			break
		}
		toMoveSelected = append(toMoveSelected, pi)
	}
	// First try removing those torrents from Transmission.
	if len(toMoveSelected) > 0 {
		logMsg := "Removing torrents from transmission"
		var toMoveTorrents []*tr.Torrent
		for _, pi := range toMoveSelected {
			logMsg += fmt.Sprintf(", %s (Magnet: %s)", pi.Name, pi.Torrent.MagnetLink)
			toMoveTorrents = append(toMoveTorrents, pi.Torrent)
		}
		a.Log("assist", logMsg)
		err := a.msv.t.RemoveTorrents(toMoveTorrents)
		if err != nil {
			return fmt.Errorf("Removing torrents from transmission failed: %v", err)
		}
		hadMoveErrors := false
		for _, pi := range toMoveSelected {
			a.Log("assist", fmt.Sprintf("Moving %s to %s", pi.Name, a.moveTarget))
			err := a.msv.moveLocked(pi, a.moveTarget)
			if err != nil {
				hadMoveErrors = true
				a.Log("assist", fmt.Sprintf("Moving %s to %s failed: %v", pi.Name, a.moveTarget, err))
			}
		}
		if hadMoveErrors {
			return fmt.Errorf("Requesting move had failures")
		}
	}

	// Enable more torrents.
	var toStart []*tr.Torrent
	for _, pi := range todo {
		if len(toStart)+len(downloading) >= a.maxConcurrentDownloading {
			break
		}
		toStart = append(toStart, pi.Torrent)
		a.Log("Starting torrent %s", pi.Name)
	}
	err := a.msv.t.StartTorrents(toStart)
	if err != nil {
		return fmt.Errorf("Failed to start torrents: %v", err)
	} else {
	}

	return nil
}

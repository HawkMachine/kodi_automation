package moveserver

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/HawkMachine/kodi_automation/platform/cron"

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

type TorrentStatus struct {
	Name        string
	MoveStatus  string
	StartStatus string
	Status      string
}

type Assistant struct {
	msv                      *MoveServer
	sleep                    time.Duration
	dryRun                   bool
	moveTarget               string
	maxConcurrentDownloading int
	maxConcurrentMoving      int

	enabled       bool
	runStarted    bool
	TorrentStatus map[string]*TorrentStatus

	j *cron.CronJob

	lock sync.Mutex
}

func newAssistant(msv *MoveServer, moveTarget string) *Assistant {
	return &Assistant{
		msv:                      msv,
		sleep:                    time.Minute,
		moveTarget:               moveTarget,
		maxConcurrentDownloading: 5,
		maxConcurrentMoving:      1,
		TorrentStatus:            map[string]*TorrentStatus{},
	}
}

func (a *Assistant) Log(tp, msg string) {
	a.msv.Log(fmt.Sprintf("Assistant.%s", tp), msg)
}

func (a *Assistant) Enable() {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.sleep < time.Minute {
		a.sleep = time.Minute
	}

	if a.j == nil {
		// TODO: do not ignore errors
		a.j, _ = a.msv.p.Cron.Register("assistant", a.assist, a.sleep)
	}
	a.j.Enable()
}

func (a *Assistant) Disable() {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.j == nil {
		return
	}
	a.j.Disable()
}

func (a *Assistant) IsEnabled() bool {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.j == nil {
		return false
	}
	return a.j.IsEnabled()
}

func (a *Assistant) shouldMove(pi *PathInfo) (bool, string) {
	if !pi.AllowMove {
		return false, "AllowMove = false"
	}
	if !pi.AllowAssistant {
		return false, "AllowAssistant = false"
	}
	if pi.MoveInfo.Moving {
		return false, "Currently moving"
	}
	if pi.Path == "" {
		return false, "Path is empty, probably just in transmission"
	}
	if pi.MoveInfo.LastError != nil {
		return false, fmt.Sprintf("Last move error is: %v", pi.MoveInfo.LastError)
	}
	if pi.Torrent == nil {
		return false, fmt.Sprintf("Torrent is nil, probably just on disk")
	}
	if pi.Torrent.DoneDate == 0 {
		return false, "Torrent.DoneDate is 0"
	}
	if pi.Torrent.PercentDone != 1.0 {
		return false, "Torrent.PercentDone != 1.0"
	}
	if pi.Torrent.Status != tr.TR_STATUS_PAUSED {
		return false, fmt.Sprintf("Torrent.Status != TR_STATUS_PAUSED, == %v", pi.Torrent.Status)
	}
	return true, "Allowed to be moved"
}

func (a *Assistant) shouldStart(pi *PathInfo) (bool, string) {
	if pi.Torrent == nil {
		return false, "Torrent is nil, probably only found on disk"
	}
	if pi.Torrent.Status != tr.TR_STATUS_PAUSED {
		return false, "Torrent.Status is not PAUSED"
	}
	if pi.Torrent.DoneDate != 0 {
		return false, fmt.Sprintf("Torrent.DoneDate != 0, == %v", pi.Torrent.DoneDate)
	}
	if pi.Torrent.PercentDone == 1.0 {
		return false, "Torrent.PercentDone == 1.0, considered done"
	}
	return true, "Allowed to start"
}

func (a *Assistant) assist() error {
	a.msv.lock.Lock()
	defer a.msv.lock.Unlock()

	// Select candidates for moving.
	// Very, very simple logic. Check if there are any torrents, allowed to move
	// without move error. These must have torrent info and be paused (I rely
	// here on a cron job that automatically pause finished torrents.
	tss := map[string]*TorrentStatus{}
	pis := a.msv.pathInfo
	toMove := []*PathInfo{}
	todo := []*PathInfo{}

	for _, pi := range pis {
		ts := &TorrentStatus{Name: pi.Name}
		tss[pi.Name] = ts
		shouldMove, moveStatus := a.shouldMove(pi)
		shouldStart, startStatus := a.shouldStart(pi)
		ts.MoveStatus = moveStatus
		ts.StartStatus = startStatus
		if shouldMove && shouldStart {
			ts.Status = "Internal error, move and start allowed at the same time."
			continue
		} else if shouldMove {
			toMove = append(toMove, pi)
		} else if shouldStart {
			todo = append(todo, pi)
		}
	}
	a.TorrentStatus = tss
	log.Printf("New tss: %v", tss)

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

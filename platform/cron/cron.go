package cron

import (
	"fmt"
	"sync"
	"time"
)

type RunInfo struct {
	Start   time.Time
	End     time.Time
	Skipped bool
	Err     error
}

type CronFunc func() error

type CronJob struct {
	name     string
	f        CronFunc
	interval time.Duration
	enabled  bool
	info     *RunInfo
	history  []*RunInfo

	lock sync.Mutex
}

func (cj *CronJob) newRun() bool {
	cj.lock.Lock()
	defer cj.lock.Unlock()

	if !cj.enabled {
		cj.history = append(cj.history, &RunInfo{Start: time.Now(), Skipped: true})
		return false
	}
	cj.info = &RunInfo{Start: time.Now()}
	return true
}

func (cj *CronJob) runEnded(err error) {
	cj.lock.Lock()
	defer cj.lock.Unlock()

	cj.info.End = time.Now()
	cj.info.Err = err

	cj.history = append(cj.history, cj.info)
	cj.info = nil
}

func (cj *CronJob) Run() {
	if !cj.newRun() {
		return
	}
	cj.runEnded(cj.f())
}

type Cron struct {
	jobs map[string]*CronJob

	lock sync.Mutex
}

func (c *Cron) Register(name string, f CronFunc, interval time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.jobs[name]; ok {
		return fmt.Errorf("Cron job %s already exists!", name)
	}

	j := &CronJob{
		name:     name,
		f:        f,
		enabled:  true,
		interval: interval,
	}
	c.jobs[name] = j
	go c.run(j)
	return nil
}

func (c *Cron) run(j *CronJob) {
	for {
		time.Sleep(j.interval)
		j.Run()
	}
}

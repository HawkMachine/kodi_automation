package cron

import (
	"fmt"
	"sync"
	"time"
)

type RunInfo struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration
	Skipped  bool
	Err      error
}

type CronFunc func() error

type CronJob struct {
	Name     string
	f        CronFunc
	Interval time.Duration
	Enabled  bool
	Info     *RunInfo
	History  []*RunInfo

	lock sync.Mutex
}

func (cj *CronJob) newRun() bool {
	cj.lock.Lock()
	defer cj.lock.Unlock()

	if !cj.Enabled {
		cj.History = append(cj.History, &RunInfo{Start: time.Now(), Skipped: true})
		return false
	}
	cj.Info = &RunInfo{Start: time.Now()}
	return true
}

func (cj *CronJob) runEnded(err error) {
	cj.lock.Lock()
	defer cj.lock.Unlock()

	cj.Info.End = time.Now()
	cj.Info.Duration = cj.Info.End.Sub(cj.Info.Start)
	cj.Info.Err = err

	if len(cj.History) > 1000 {
		cj.History = cj.History[:1000]
	}
	cj.History = append([]*RunInfo{cj.Info}, cj.History...)
	cj.Info = nil
}

func (cj *CronJob) Run() {
	if !cj.newRun() {
		return
	}
	cj.runEnded(cj.f())
}

func (cj *CronJob) Enable() {
	cj.lock.Lock()
	defer cj.lock.Unlock()

	cj.Enabled = true
}

func (cj *CronJob) Disable() {
	cj.lock.Lock()
	defer cj.lock.Unlock()

	cj.Enabled = false
}

func (cj *CronJob) IsEnabled() bool {
	cj.lock.Lock()
	defer cj.lock.Unlock()

	return cj.Enabled
}

type Cron struct {
	jobs map[string]*CronJob

	lock sync.Mutex
}

func New() *Cron {
	return &Cron{
		jobs: map[string]*CronJob{},
	}
}

func (c *Cron) CronJobs() map[string]*CronJob {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.jobs
}

func (c *Cron) Register(name string, f CronFunc, interval time.Duration) (*CronJob, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.jobs[name]; ok {
		return nil, fmt.Errorf("Cron job %s already exists!", name)
	}

	j := &CronJob{
		Name:     name,
		f:        f,
		Enabled:  true,
		Interval: interval,
	}
	c.jobs[name] = j
	go c.run(j)
	return j, nil
}

func (c *Cron) run(j *CronJob) {
	for {
		time.Sleep(j.Interval)
		j.Run()
	}
}

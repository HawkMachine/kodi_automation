package cronview

import (
	"net/http"

	"github.com/HawkMachine/kodi_automation/platform"
	"github.com/HawkMachine/kodi_automation/platform/cron"
	"github.com/HawkMachine/kodi_automation/server"
)

type CronView struct {
	p *platform.Platform
}

func (cv *CronView) GetName() string {
	return "cronview"
}

func (cv *CronView) GetTemplates() map[string][]string {
	return map[string][]string{
		"cron_page": []string{
			"base.html",
			"cron.html",
		},
	}
}

func (cv *CronView) GetHandlers() map[string]server.ViewHandle {
	return map[string]server.ViewHandle{
		"/cron": server.NewViewHandle(cv.cronPage),
	}
}

func (cv *CronView) GetMenu() (string, map[string]string) {
	return "Cron", map[string]string{
		"Cron": "/cron",
	}
}

func New(server server.HTTPServer, p *platform.Platform) server.View {
	return &CronView{p: p}
}

func (cv *CronView) cronPage(w http.ResponseWriter, r *http.Request, s server.HTTPServer) {
	context := struct {
		CronJobs map[string]*cron.CronJob
	}{
		CronJobs: cv.p.Cron.CronJobs(),
	}
	s.RenderTemplate(w, r, cv.GetName(), "cron_page", "Cron", context)
}

package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/HawkMachine/kodi_automation/server/auth"
)

// Interface used by ViewHandler
type HTTPServer interface {
	RenderTemplate(w http.ResponseWriter, r *http.Request, viewName, templateName string, title string, context interface{})
}

// Type representing a particular handler for a View.
type ViewHandle interface {

	// Called by server upon receving the request. Returns context used for
	// rendering the web page or error.
	ServeHTTP(http.ResponseWriter, *http.Request, HTTPServer)
}

// Represents a portion of view logic for some stuff on the server. Usually
// contains common data used by multiple view handlers for that view.
type View interface {
	// Returns View name.
	GetName() string

	// Returns a mapping from template name to list of filenames under template
	// directory. These templates will be parsed by server at startup and can be
	// used by ViewHandle with RenderTemplate method of MyHTTPServer.
	GetTemplates() map[string][]string

	// Returns a mapping from url to ViewHandle that handles that URL.
	GetHandlers() map[string]ViewHandle

	// Returns a map from name to url. For each an entry will be created in
	// navigation pane. If the map is empty, no entries will be added.
	GetMenu() (string, map[string]string)
}

// Implementation of ViewHandle that is using a function.
type viewHandleFunc struct {
	f func(http.ResponseWriter, *http.Request, HTTPServer)
}

func (vhf *viewHandleFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, s HTTPServer) {
	vhf.f(w, r, s)
}

func NewViewHandle(f func(http.ResponseWriter, *http.Request, HTTPServer)) ViewHandle {
	return &viewHandleFunc{f: f}
}

type basePageContext struct {
	Title          string
	IsMobile       bool
	ContentContext interface{}
	Links          map[string]string
	IframeLinks    map[string]string
	ViewsMenu      map[string]map[string]string
	Errors         []string
}

type MyHTTPServer struct {
	views           map[string]View
	parsedTemplates map[string]map[string]*template.Template

	port int

	links       map[string]string
	iframeLinks map[string]string

	templatesPath string
	resourcesPath string

	basicAuthUsername string
	basicAuthPassword string

	templateFuncs template.FuncMap
}

func (s *MyHTTPServer) RenderTemplate(w http.ResponseWriter, r *http.Request, viewName, templateName string, title string, context interface{}) {
	t, err := s.getTemplate(viewName, templateName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	baseContext := s.getBaseContext(r)
	baseContext.ContentContext = context
	baseContext.Title = title
	err = t.ExecuteTemplate(w, "base", baseContext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *MyHTTPServer) BasicAuthAuthenticate(username, password string) (bool, error) {
	log.Printf("AUTH: %s : %s\n", username, password)
	if s.basicAuthPassword == "" {
		return true, nil
	}
	return s.basicAuthUsername == username && s.basicAuthPassword == password, nil
}

func logHandleWrap(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("REQ : %s : %#v\n", r.RemoteAddr, r)
		f(w, r)
	}
}

// makeHTTPHandler returns a function that can be used directly to register a
// handler for a url with http.HandleFunc
func makeHTTPHandleFunc(s *MyHTTPServer, h ViewHandle) func(http.ResponseWriter, *http.Request) {
	return logHandleWrap(auth.BasicAuthWrap(s, func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r, s)
	}))
}

// RegisterView registers the given View on the server.
func (s *MyHTTPServer) RegisterView(v View) error {
	log.Printf("Registering view %q\n", v.GetName())
	_, ok := s.parsedTemplates[v.GetName()]
	if ok {
		return fmt.Errorf("view with name %s is already registered", v.GetName())
	}

	s.views[v.GetName()] = v

	// Parse view templates.
	viewTemplates := map[string]*template.Template{}
	for name, relPaths := range v.GetTemplates() {
		log.Printf("  registering template %q: %v\n", name, relPaths)
		var paths []string
		for _, relPath := range relPaths {
			paths = append(paths, filepath.Join(s.templatesPath, relPath))
		}
		t := template.Must(template.New(name).Funcs(s.templateFuncs).ParseFiles(paths...))
		viewTemplates[name] = t
	}
	s.parsedTemplates[v.GetName()] = viewTemplates

	// Register handlers.
	for url, h := range v.GetHandlers() {
		log.Printf("  registering url %q\n", url)
		http.HandleFunc(url, makeHTTPHandleFunc(s, h))
	}
	return nil
}

func (s *MyHTTPServer) Run() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir(s.resourcesPath))))

	bindAddr := fmt.Sprintf(":%d", s.port)
	log.Printf("Bind address %s", bindAddr)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}

func (s *MyHTTPServer) getTemplate(viewName, templateName string) (*template.Template, error) {
	viewTemplates, ok := s.parsedTemplates[viewName]
	if !ok {
		return nil, fmt.Errorf("Unknown view %s", viewName)
	}

	t, ok := viewTemplates[templateName]
	if !ok {
		return nil, fmt.Errorf("Unknown template %s for %s view", templateName, viewName)
	}
	return t, nil
}

func (s *MyHTTPServer) getBaseContext(r *http.Request) basePageContext {
	return basePageContext{
		IsMobile:    s.isMobile(r),
		Links:       s.links,
		IframeLinks: s.iframeLinks,
		ViewsMenu:   s.getViewsMenu(),
	}
}

func (s *MyHTTPServer) getViewsMenu() map[string]map[string]string {
	menu := map[string]map[string]string{}
	for _, v := range s.views {
		title, vmenu := v.GetMenu()
		if len(vmenu) != 0 {
			menu[title] = vmenu
		}
	}
	log.Printf("Menu: %v\n", menu)
	return menu
}

func (s *MyHTTPServer) isMobile(r *http.Request) bool {
	// I'm so lazy...
	return strings.Contains(r.UserAgent(), "Nexus")
}

// New creates new instance of MyHTTPServer.
func NewMyHTTPServer(port int, basicAuthUsername string, basicAuthPassword string,
	templatesPath string, resourcesPath string,
	links map[string]string, iframeLinks map[string]string) *MyHTTPServer {

	if links == nil {
		links = map[string]string{}
	}
	if iframeLinks == nil {
		iframeLinks = map[string]string{}
	}
	httpServer := &MyHTTPServer{
		views:           map[string]View{},
		parsedTemplates: map[string]map[string]*template.Template{},

		port:        port,
		links:       links,
		iframeLinks: iframeLinks,

		templatesPath: templatesPath,
		resourcesPath: resourcesPath,

		basicAuthUsername: basicAuthUsername,
		basicAuthPassword: basicAuthPassword,

		templateFuncs: template.FuncMap{
			"timeformat": func(v time.Time, f string) string {
				if f == "" {
					f = "Mon Jan 2 15:04"
				}
				return v.Format(f)
			},
			"sizeformat": func(size int64) string {
				if size < 1000 {
					return fmt.Sprintf("%dB", size)
				} else if size < 1000000 {
					return fmt.Sprintf("%.2fKB", float64(size)/1000.0)
				} else if size < 10^9 {
					return fmt.Sprintf("%.2fMB", float64(size)/1000000.0)
				}
				return fmt.Sprintf("%.2fGB", float64(size)/1000000000.0)
			},
		},
	}
	return httpServer
}

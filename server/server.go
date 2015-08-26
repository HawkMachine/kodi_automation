package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// Interface used by ViewHandler
type HTTPServer interface {
	RenderTemplate(w http.ResponseWriter, r *http.Request, viewName, templateName string, context interface{})
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
	Errors         []string
}

type MyHTTPServer struct {
	parsedTemplates map[string]map[string]*template.Template

	port int

	links       map[string]string
	iframeLinks map[string]string

	templatesPath string
}

func (s *MyHTTPServer) RenderTemplate(w http.ResponseWriter, r *http.Request, viewName, templateName string, context interface{}) {
	t, err := s.getTemplate(viewName, templateName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	baseContext := s.getBaseContext(r)
	baseContext.ContentContext = context
	err = t.ExecuteTemplate(w, "base", baseContext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// makeHTTPHandler returns a function that can be used directly to register a
// handler for a url with http.HandleFunc
func makeHTTPHandleFunc(s HTTPServer, h ViewHandle) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r, s)
	}
}

// RegisterView registers the given View on the server.
func (s *MyHTTPServer) RegisterView(v View) error {
	log.Printf("Registering view %q\n", v.GetName())
	_, ok := s.parsedTemplates[v.GetName()]
	if ok {
		return fmt.Errorf("view with name %s is already registered", v.GetName())
	}

	// Parse view templates.
	viewTemplates := map[string]*template.Template{}
	for name, relPaths := range v.GetTemplates() {
		log.Printf("  registering template %q: %v\n", name, relPaths)
		var paths []string
		for _, relPath := range relPaths {
			paths = append(paths, filepath.Join(s.templatesPath, relPath))
		}
		t := template.Must(template.New(name).ParseFiles(paths...))
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
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

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
	}
}

func (s *MyHTTPServer) isMobile(r *http.Request) bool {
	// I'm so lazy...
	return strings.Contains(r.UserAgent(), "Nexus")
}

// New creates new instance of MyHTTPServer.
func NewMyHTTPServer(port int, templatesPath string, resourcesPath string, links map[string]string, iframeLinks map[string]string) *MyHTTPServer {
	if links == nil {
		links = map[string]string{}
	}
	if iframeLinks == nil {
		iframeLinks = map[string]string{}
	}
	httpServer := &MyHTTPServer{
		parsedTemplates: map[string]map[string]*template.Template{},

		port:          port,
		links:         links,
		iframeLinks:   iframeLinks,
		templatesPath: templatesPath,
	}
	return httpServer
}

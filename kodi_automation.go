package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/HawkMachine/kodi_automation/moveserver"
	"github.com/HawkMachine/kodi_automation/platform"
	"github.com/HawkMachine/kodi_automation/server"
	"github.com/HawkMachine/kodi_automation/views/cronview"
	"github.com/HawkMachine/kodi_automation/views/kodiview"
	"github.com/HawkMachine/kodi_automation/views/moveserverview"
	"github.com/HawkMachine/kodi_automation/views/transmissionview"
	"github.com/HawkMachine/kodi_automation/views/uploadtorrentview"
	"github.com/HawkMachine/kodi_automation/views/wrapview"
)

var (
	mvBufferSize      = flag.Int("mv_buffer_size", 5, "size of the mv request buffer")
	maxMvCommands     = flag.Int("max_mv_commands", 5, "max mv commands running in parallel")
	port              = flag.Int("port", 8080, "port to use")
	sourceDir         = flag.String("source_dir", "", "directory to scan")
	moviesTarget      = flag.String("movies_dir", "", "comma-separated list of directories where to move movies")
	seriesTarget      = flag.String("series_dir", "", "where to move series")
	customLinks       = flag.String("links", "", "comma-delimited list of <link name>:<url>")
	customIframeLinks = flag.String("iframe_links", "", "comma-delimited list of <link name>:<url>")
	waitForIP         = flag.Int("wait_for_ip", 300, "Seconds to wait for IP address")

	templatesPath = flag.String("templates_path", "templates", "Path to server templates.")
	resourcesPath = flag.String("resources_path", "resources", "Path to server resources.")

	kodiAddress  = flag.String("kodi_address", "", "Address of kodi instance")
	kodiUsername = flag.String("kodi_username", "", "Username of kodi user to use")
	kodiPassword = flag.String("kodi_password", "", "Password of kodi user to use")

	transmissionAddress  = flag.String("transmission_address", "", "Address of transmission.")
	transmissionUsername = flag.String("transmission_username", "", "Username of transmission.")
	transmissionPassword = flag.String("transmission_password", "", "Password of transmission.")

	basicAuthUsername = flag.String("auth_username", "admin", "Basic auth username")
	basicAuthPassword = flag.String("auth_password", "4dm1n", "Basic auth password")

	configFile = flag.String("config_file", "", "Config file")
)

type config struct {
	Port int `json:"port,omitempty"`

	MoveServer moveserver.MoveServerConfig `json:"move_server,omitempty"`

	Links       map[string]string `json:"links,omitempty"`
	IframeLinks map[string]string `json:"iframe_links,omitempty"`

	TemplatesPath string `json:"templates_paths,omitempty"`
	ResourcesPath string `json:"resources_paths,omitempty"`

	KodiAddress  string `json:"kodi_address,omitempty"`
	KodiUsername string `json:"kodi_username,omitempty"`
	KodiPassword string `json:"kodi_password,omitempty"`

	TransmissionAddress  string `json:"transmission_address,omitempty"`
	TransmissionUsername string `json:"transmission_username,omitempty"`
	TransmissionPassword string `json:"transmission_password,omitempty"`

	BasicAuthUsername string `json:"basic_auth_username,omitempty"`
	BasicAuthPassword string `json:"basic_auth_password,omitempty"`

	WaitForIP int `json:"wait_for_ip,omitempty"`
}

func loadConfigFromFile(path string) (*config, error) {
	bts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg *config = &config{}
	err = json.Unmarshal(bts, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// getLocalIP returns the non loopback local IP of the host.
// http://stackoverflow.com/a/31551220
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}

func parseCustomLinksFlag(str string) map[string]string {
	res := map[string]string{}
	for _, item := range strings.Split(str, ",") {
		if item == "" {
			continue
		}
		linkItems := strings.SplitN(item, ":", 2)
		if len(linkItems) < 2 {
			log.Printf("Wrong flag format: ", item)
			continue
		}
		res[linkItems[0]] = linkItems[1]
	}
	return res
}

func replaceLocalHost(m map[string]string, ip string) map[string]string {
	n := map[string]string{}
	for k, v := range m {
		n[k] = strings.Replace(v, "localhost", ip, 1)
	}
	return n
}

func main() {
	flag.Parse()

	var cfg *config

	if *configFile != "" {
		var err error
		log.Printf("Loading from config %s", *configFile)
		cfg, err = loadConfigFromFile(*configFile)
		log.Println(cfg)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		links := parseCustomLinksFlag(*customLinks)
		iframeLinks := parseCustomLinksFlag(*customIframeLinks)

		cfg = &config{
			Port: *port,
			MoveServer: moveserver.MoveServerConfig{
				SourceDir:     *sourceDir,
				SeriesTargets: strings.Split(*seriesTarget, ","),
				MoviesTargets: strings.Split(*moviesTarget, ","),
				MvBufferSize:  *mvBufferSize,
				MaxMvCommands: *maxMvCommands,
			},

			TemplatesPath: *templatesPath,
			ResourcesPath: *resourcesPath,

			KodiAddress:  *kodiAddress,
			KodiUsername: *kodiUsername,
			KodiPassword: *kodiPassword,

			TransmissionAddress:  *transmissionAddress,
			TransmissionUsername: *transmissionUsername,
			TransmissionPassword: *transmissionPassword,

			Links:       links,
			IframeLinks: iframeLinks,

			BasicAuthUsername: *basicAuthUsername,
			BasicAuthPassword: *basicAuthPassword,

			WaitForIP: *waitForIP,
		}
	}

	log.Printf("CONFIG           = %#v", cfg)

	var err error
	if cfg.MoveServer.SourceDir == "" {
		log.Fatal("Missing source directory")
	}
	if cfg.MoveServer.SourceDir, err = filepath.Abs(cfg.MoveServer.SourceDir); err != nil {
		log.Fatal(err)
	}
	if len(cfg.MoveServer.MoviesTargets) == 0 {
		log.Fatal(fmt.Errorf("Movies targets resolves to empty list"))
	}
	if len(cfg.MoveServer.SeriesTargets) == 0 {
		log.Fatal(fmt.Errorf("Series targets resolves to empty list"))
	}

	if cfg.MoveServer.MaxMvCommands <= 5 {
		cfg.MoveServer.MaxMvCommands = 5
	}
	if cfg.MoveServer.MvBufferSize <= 5 {
		cfg.MoveServer.MvBufferSize = 5
	}
	if cfg.WaitForIP <= 5 {
		cfg.WaitForIP = 5
	}

	// Trying to get IP.
	var ip string
	start := time.Now()
	deadline := start.Add(time.Duration(cfg.WaitForIP) * time.Second)
	for {
		if time.Now().After(deadline) {
			log.Fatalf("Could not get local IP for %d seconds", cfg.WaitForIP)
		}
		ip, err = getLocalIP()
		if err != nil {
			fmt.Printf("Got error while resolving IP: %v", err)
		} else if ip != "" {
			break
		} else {
			log.Printf("Resolved IP to empty string, %s left", deadline.Sub(time.Now()).String())
		}
		time.Sleep(time.Second * 5)
	}

	if !strings.HasSuffix(cfg.KodiAddress, "/jsonrpc") {
		cfg.KodiAddress += "/jsonrpc"
	}

	cfg.Links = replaceLocalHost(cfg.Links, ip)
	cfg.IframeLinks = replaceLocalHost(cfg.IframeLinks, ip)

	log.Printf("CONFIG           = %#v", cfg)
	log.Printf("PORT             = %d", cfg.Port)
	log.Printf("TEMPLATES PATH   = %s", cfg.TemplatesPath)
	log.Printf("RESOURCES PATH   = %s", cfg.ResourcesPath)
	log.Printf("SOURCE DIR       = %s", cfg.MoveServer.SourceDir)
	log.Printf("MOVIES TARGETS   = %v", cfg.MoveServer.MoviesTargets)
	log.Printf("SERIES TARGETS   = %s", cfg.MoveServer.SeriesTargets)
	log.Printf("MAX MV COMMANDS  = %d", cfg.MoveServer.MaxMvCommands)
	log.Printf("MV BUFFER SIZE   = %d", cfg.MoveServer.MvBufferSize)
	log.Printf("WAIT FOR IP      = %d", cfg.WaitForIP)
	log.Printf("LINKS            = %v", cfg.Links)
	log.Printf("IFRAME LINKS     = %v", cfg.IframeLinks)
	log.Printf("LINKS            = %v", cfg.Links)
	log.Printf("IFRAME LINKS     = %v", cfg.IframeLinks)
	log.Printf("IP               = %s", ip)

	c := platform.NewConfigFromStrings(
		cfg.KodiAddress,
		cfg.KodiUsername,
		cfg.KodiPassword,
		cfg.TransmissionAddress,
		cfg.TransmissionUsername,
		cfg.TransmissionPassword)

	p := platform.New(c)

	// Server.
	s := server.NewMyHTTPServer(
		cfg.Port,
		cfg.BasicAuthUsername,
		cfg.BasicAuthPassword,
		cfg.TemplatesPath,
		cfg.ResourcesPath,
		cfg.Links,
		cfg.IframeLinks)

	// Initialize move server view.
	moveServer, err := moveserver.New(p, cfg.MoveServer)
	if err != nil {
		log.Fatal(err)
	}

	views := []server.View{}

	moveServerView, err := moveserverview.New(s, moveServer)
	if err != nil {
		log.Fatal(err)
	}
	views = append(views, moveServerView)

	// Cron View
	views = append(views, cronview.New(s, p))

	// Wrap view.
	views = append(views, wrapview.New(cfg.IframeLinks))

	// Kodi stats view.
	if cfg.KodiAddress != "" {
		var scanTargets []string
		scanTargets = append(scanTargets, cfg.MoveServer.MoviesTargets...)
		scanTargets = append(scanTargets, cfg.MoveServer.SeriesTargets...)
		views = append(views, kodiview.New(p, scanTargets))
	} else {
		log.Println("Kodi address missing. Skipping kodi stats view.")
	}

	if cfg.TransmissionAddress != "" {
		// Transmission view
		views = append(views, transmissionview.New(p))

		// Upload torrent view
		if utv, err := uploadtorrentview.New(p, moveServer); err != nil {
			views = append(views, utv)
		}
	}

	// Run server.
	for _, v := range views {
		err = s.RegisterView(v)
		if err != nil {
			log.Fatal("Error occured while registering view %q: %v\n", v.GetName(), err)
		}
	}

	s.Run()
}

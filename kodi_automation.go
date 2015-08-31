package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/HawkMachine/kodi_automation/moveserver"
	"github.com/HawkMachine/kodi_automation/server"
	"github.com/HawkMachine/kodi_automation/views/kodiview"
	"github.com/HawkMachine/kodi_automation/views/moveserverview"
	"github.com/HawkMachine/kodi_automation/views/wrapview"
)

var (
	mvBufferSize      = flag.Int("mv_buffer_size", 5, "size of the mv request buffer")
	maxMvCommands     = flag.Int("max_mv_commands", 5, "max mv commands running in parallel")
	port              = flag.Int("port", 8080, "port to use")
	sourceDir         = flag.String("source_dir", "", "directory to scan")
	moviesTarget      = flag.String("movies_dir", "", "where to move movies")
	seriesTarget      = flag.String("series_dir", "", "where to move series")
	customLinks       = flag.String("links", "", "comma-delimited list of <link name>:<url>")
	customIframeLinks = flag.String("iframe_links", "", "comma-delimited list of <link name>:<url>")
	waitForIP         = flag.Int("wait_for_ip", 300, "Seconds to wait for IP address")

	templatesPath = flag.String("templates_path", "templates", "Path to server templates.")
	resourcesPath = flag.String("resources_path", "resources", "Path to server resources.")

	kodiAddress  = flag.String("kodi_address", "", "Address of kodi instance")
	kodiUsername = flag.String("kodi_username", "", "Username of kodi user to use")
	kodiPassword = flag.String("kodi_password", "", "Password of kodi user to use")

	basicAuthUsername = flag.String("auth_username", "admin", "Basic auth username")
	basicAuthPassword = flag.String("auth_password", "4dm1n", "Basic auth password")
)

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
	var err error
	if *sourceDir == "" || *moviesTarget == "" || *seriesTarget == "" {
		log.Fatal("Missing source, series or movies target directory")
	}

	if *sourceDir, err = filepath.Abs(*sourceDir); err != nil {
		log.Fatal(err)
	}
	if *moviesTarget, err = filepath.Abs(*moviesTarget); err != nil {
		log.Fatal(err)
	}
	if *seriesTarget, err = filepath.Abs(*seriesTarget); err != nil {
		log.Fatal(err)
	}
	if *maxMvCommands <= 5 {
		*maxMvCommands = 5
	}
	if *mvBufferSize <= 5 {
		*mvBufferSize = 5
	}
	log.Printf("PORT             = %d", *port)
	log.Printf("SOURCE_DIR       = %s", *sourceDir)
	log.Printf("MOVIES_TARGET    = %s", *moviesTarget)
	log.Printf("SERIES_TARGET    = %s", *seriesTarget)
	log.Printf("MAX_MV_COMMANDS  = %d", *maxMvCommands)
	log.Printf("MV_BUFFER_SIZE   = %d", *mvBufferSize)
	log.Printf("WAIT_FOR_IP      = %d", *waitForIP)
	log.Printf("LINKS            = %v", *customLinks)
	log.Printf("IFRAME LINKS     = %v", *customIframeLinks)

	var ip string
	start := time.Now()
	deadline := start.Add(time.Duration(*waitForIP) * time.Second)
	for {
		if time.Now().After(deadline) {
			log.Fatalf("Could not get local IP for %d seconds", *waitForIP)
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

	log.Printf("IP               = %s", ip)
	links := replaceLocalHost(parseCustomLinksFlag(*customLinks), ip)
	iframeLinks := replaceLocalHost(parseCustomLinksFlag(*customIframeLinks), ip)
	log.Printf("LINKS            = %v", links)
	log.Printf("IFRAME LINKS     = %v", iframeLinks)

	// Server.
	s := server.NewMyHTTPServer(
		*port,
		*basicAuthUsername,
		*basicAuthPassword,
		*templatesPath,
		*resourcesPath,
		links,
		iframeLinks)

	// Initialize move server view.
	moveServer, err := moveserver.New(
		*sourceDir, *moviesTarget, *seriesTarget, *maxMvCommands, *mvBufferSize)
	if err != nil {
		log.Fatal(err)
	}

	views := []server.View{}

	moveServerView, err := moveserverview.New(s, moveServer)
	if err != nil {
		log.Fatal(err)
	}
	views = append(views, moveServerView)

	// Wrap view.
	views = append(views, wrapview.New(iframeLinks))

	// Kodi stats view.
	if *kodiAddress != "" && *kodiUsername != "" && *kodiPassword != "" {
		views = append(views, kodiview.New(*kodiAddress, *kodiUsername, *kodiPassword))
	} else {
		log.Println("Kodi address, username or password missing. Skipping kodi stats view.")
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

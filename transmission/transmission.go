package transmission

// Transmission api as described in https://trac.transmissionbt.com/browser/trunk/extras/rpc-spec.txt

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	//	"encoding/json"
	//	"math/rand"
	"net/http"
)

var (
	DOWNLOAD_DIR = "/home/kuba/Videos"
)

type Request struct {
	Method    string            // required
	Arguments map[string]string // optional
	Tag       string            // optional
}

type Transmission struct {
	hostname string
	port     int
	user     string
	password string
	url      string
}

func New(hostname string, port int, user string, password string) *Transmission {
	t := new(Transmission)
	t.hostname = hostname
	t.port = port
	t.user = user
	t.password = password
	t.url = "http://" + user + ":" + password + "@" + hostname + ":" + strconv.Itoa(port) + "/transmission/rpc"
	t.url = "http://" + hostname + ":" + strconv.Itoa(port) + "/transmission/rpc/"
	return t
}

type torrentGetArguments struct {
	Fields []string `json:"fields"`
	Ids    []int    `json:"ids"`
}

type torrentGetRequest struct {
	Method    string              `json:"method"`
	Arguments torrentGetArguments `json:"arguments"`
	Tag       int64               `json:"tag"`
}

type TorrentFileInfo struct {
	Name           string `json:"name"`
	BytesCompleted int64  `json:"bytesCompleted"`
	Length         int64  `json:"length"`
}

type TorrentInfo struct {
	AddedDate   int64
	DoneDate    int64
	DownloadDir string
	Error       int
	ErrorString string
	Eta         int64
	Id          int
	IsFinished  bool
	Files       []TorrentFileInfo
	Name        string
	PercentDone float64
	Status      string
}

type torrentsInfo struct {
	arguments struct {
		torrents []TorrentInfo
	}
}

func (t *Transmission) getTorrentGetRequest(ids []int) torrentGetRequest {
	// tag := rand.Int63n(100000)
	r := torrentGetRequest{
		Method: "torrent-get",
		Arguments: torrentGetArguments{
			Fields: []string{
				"addedDate",
				"doneDate",
				"downloadDir",
				"error",
				"errorString",
				"eta",
				"id",
				"isFinished",
				"files",
				"name",
				"percentDone",
				"status",
			},
		},
		// Tag: tag,
		Tag: 4,
	}
	if ids != nil {
		r.Arguments.Ids = ids
	}
	return r
}

func (t *Transmission) getHttpRequest(body []byte, sessionId string) (*http.Request, error) {
	// Make the http request.
	// TODO - do not pass the body as body but as formatted form.
	httpReq, err := http.NewRequest("POST", t.url, bytes.NewReader(body))
	if err != nil {
		return httpReq, err
	}
	httpReq.SetBasicAuth(t.user, t.password)
	if sessionId != "" {
		httpReq.Header.Set("X-Transmission-Session-Id", sessionId)
	}
	return httpReq, nil
}

func directoryListing(dirname string) ([]string, error) {
	res := []string{}
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		res = append(res, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	// fmt.Printf("%v\n\n", res)
	return res, nil
}

func (t *Transmission) GetTorrents() ([]TorrentInfo, error) {
	fis, err := ioutil.ReadDir(DOWNLOAD_DIR)
	if err != nil {
		return nil, err
	}
	res := []TorrentInfo{}
	for _, fi := range fis {
		// Get files of this torrent.
		paths := []string{}
		if fi.IsDir() {
			path := filepath.Join(DOWNLOAD_DIR, fi.Name())
			paths, err = directoryListing(path)
			if err != nil {
				return nil, err
			}
		} else {
			paths = []string{filepath.Join(DOWNLOAD_DIR, fi.Name())}
		}

		// Check if is finished.
		isFinished := true
		for _, path := range paths {
			if filepath.Ext(path) == ".part" {
				isFinished = false
				break
			}
		}
		tfi := []TorrentFileInfo{}
		for _, path := range paths {
			tfi = append(tfi, TorrentFileInfo{Name: path})
		}
		ti := TorrentInfo{
			Name:        fi.Name(),
			IsFinished:  isFinished,
			Files:       tfi,
			DownloadDir: DOWNLOAD_DIR,
		}
		res = append(res, ti)
	}
	return res, nil
}

//func (t *Transmission) GetTorrents(ids []int) ([]TorrentInfo, error) {
//  r := t.getTorrentGetRequest([]int{})
//  fmt.Printf("::: TREQ :::\n%v\n\n", r)
//
//	body, err := json.Marshal(r)
//	if err != nil {
//		return nil, err
//	}
//  fmt.Printf("::: JSON :::\n%s\n\n", body)
//
//  httpReq, err := t.getHttpRequest(body, "")
//  if err != nil {
//    return nil, err
//  }
//
//  fmt.Printf("::: REQ  :::\n%#v\n\n", httpReq)
//  cli := http.Client{}
//	resp, err := cli.PostForm(httpReq)
//	if err != nil {
//		return nil, err
//	}
//  fmt.Printf("::: RESP1 :::\n%#v\n\n", resp)
//
//  if resp.StatusCode == 409 {
//    sessionId := resp.Header.Get("X-Transmission-Session-Id")
//    if sessionId == "" {
//      return nil, fmt.Errorf("http reposnse with 409 and no session id")
//    }
//    httpReq, err = t.getHttpRequest(body, sessionId)
//    if err != nil {
//      return nil, err
//    }
//    fmt.Printf("::: REQ  :::\n%#v\n\n", httpReq)
//    resp, err = cli.Do(httpReq)
//    if err != nil {
//      return nil, err
//    }
//    fmt.Printf("::: RESP2 :::\n%v\n\n", resp)
//    fmt.Printf("::: RESP2 :::\n%#v\n\n", resp)
//  }
//
//	decoder := json.NewDecoder(resp.Body)
//	var tis torrentsInfo
//	decoder.Decode(&tis)
//	return tis.arguments.torrents, nil
//}

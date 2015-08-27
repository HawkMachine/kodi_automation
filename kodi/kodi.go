package kodi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

type request struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Id      string      `json:"id"`
	Params  interface{} `json:"params"`
}

type ErrorStack struct {
	Name     string      `json:"name,omitempty"`
	Type     string      `json:"type,omitempty"`
	Message  string      `json:"message,omitempty"`
	Property *ErrorStack `json:"property,omitempty"`
}

type ErrorData struct {
	Method string      `json:"method"`
	Stack  *ErrorStack `json:"stack"`
}

type ResponseError struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    *ErrorData `json:"data,omitempty"`
}

type ResponseBase struct {
	Jsonrpc string         `json:"jsonrpc,omitempty"`
	Method  string         `json:"method,omitempty"`
	Id      string         `json:"id,omitempty"`
	Error   *ResponseError `json:"error,omitempty"`
}

// http://kodi.wiki/view/JSON-RPC_API/v6
type Kodi struct {
	address  string
	username string
	password string

	VideoLibrary *VideoLibrary
}

func (k *Kodi) postRequest(r interface{}) (*http.Response, error) {
	bts, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	glog.V(3).Infof("KODI POST REQUEST  : %v\n", string(bts))

	cli := &http.Client{}
	req, err := http.NewRequest("POST", k.address, bytes.NewBuffer(bts))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(k.username, k.password)

	response, err := cli.Do(req)
	glog.V(3).Infof("KODI POST RESPONSE : %v\n", response)

	return response, err
}

func (k *Kodi) doRPC(request interface{}, response interface{}) error {
	resp, err := k.postRequest(request)
	if err != nil {
		return err
	}
	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	glog.V(2).Infof("KODI JSON RESPONSE : %v\n", string(bts))

	dec := json.NewDecoder(bytes.NewBuffer(bts))
	err = dec.Decode(response)
	return err
}

func New(address, username, password string) *Kodi {
	k := &Kodi{
		address:  address,
		username: username,
		password: password,
	}

	k.VideoLibrary = &VideoLibrary{k: k}
	return k
}

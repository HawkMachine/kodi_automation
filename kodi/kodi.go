package kodi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	fmt.Printf("POST : %v\n", string(bts))
	if err != nil {
		return nil, err
	}
	cli := &http.Client{}
	req, err := http.NewRequest("POST", k.address, bytes.NewBuffer(bts))
	req.SetBasicAuth(k.username, k.password)
	return cli.Do(req)
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

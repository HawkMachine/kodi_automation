package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

type BasicAuthAuthenticator interface {
	BasicAuthAuthenticate(username, password string) (bool, error)
}

func basicAuthAuthenticate(a BasicAuthAuthenticator, r *http.Request) (bool, error) {
	authorization := r.Header.Get("Authorization")
	items := strings.SplitN(authorization, " ", 2)
	if len(items) != 2 || items[0] != "Basic" {
		return false, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(items[1])
	if err != nil {
		return false, fmt.Errorf("Base64 decoding error")
	}

	authItems := strings.SplitN(string(decoded), ":", 2)
	if len(authItems) != 2 {
		return false, fmt.Errorf("Did not found \":\" in basic auth")
	}

	return a.BasicAuthAuthenticate(authItems[0], authItems[1])
}

func BasicAuthWrap(a BasicAuthAuthenticator, f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	// Basic auth implemented basing on:
	// https://en.wikipedia.org/wiki/Basic_access_authentication#Protocol

	return func(w http.ResponseWriter, r *http.Request) {
		ok, err := basicAuthAuthenticate(a, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if ok {
			f(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"test\"")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("401 Unauthorized"))
		}
	}
}

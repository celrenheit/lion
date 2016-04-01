package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/celrenheit/lion"
	"golang.org/x/net/context"
)

const basicAuthPrefix = "Basic "

var user = []byte("lion")
var pass = []byte("argh")

func home(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home")
}

func basicAuthMiddleware(next lion.Handler) lion.Handler {
	return lion.HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		if strings.HasPrefix(auth, basicAuthPrefix) {
			// Check credentials
			payload, err := base64.StdEncoding.DecodeString(auth[len(basicAuthPrefix):])
			if err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 &&
					bytes.Equal(pair[0], user) &&
					bytes.Equal(pair[1], pass) {

					// Delegate request to the given handle
					next.ServeHTTPC(c, w, r)
					return
				}
			}
		}

		// Request Basic Authentication otherwise
		w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

func protectedHome(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Connected to the protected home")
}

func main() {
	l := lion.Classic()
	l.GetFunc("/", home)

	g := l.Group("/protected")
	g.UseFunc(basicAuthMiddleware)
	g.GetFunc("/", protectedHome)

	l.Run(":3000")
}

package main

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/celrenheit/lion"
)

func main() {
	l := lion.New()

	// Group by /api basepath
	api := l.Group("/api")

	// Specific to v1
	v1 := api.Subrouter().
		Host("v1.local.dev:3000")

	v1.Get("/", handler("v1"))

	// Specific to v2
	v2 := api.Subrouter().
		Host("v2.local.dev:3000")

	v2.Get("/", handler("v2"))

	// Common
	l.Get("/", handler(`
        Setup your /etc/hosts file like so:
        127.0.0.1 v1.local.dev v2.local.dev

        Make two requests to: v1.local.dev:3000/api and v2.local.dev:3000/api

        Using curl:

        $ curl v1.local.dev:3000/api
        v1 (expected output)

        $ curl v2.local.dev:3000/api
        v2 (expected output)
    `))
	l.Run()
}

func handler(name string) lion.Handler {
	return lion.HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(name))
	})
}

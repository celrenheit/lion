package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/celrenheit/lion"
	"golang.org/x/net/context"
)

func Home(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home")
}

func Hello(c context.Context, w http.ResponseWriter, r *http.Request) {
	ctx := lion.C(c)
	fmt.Fprintf(w, "Hello "+ctx.Param("name"))
}

type logger struct{}

func (*logger) ServeNext(next lion.Handler) lion.Handler {
	return lion.HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTPC(c, w, r)

		fmt.Printf("Served %s in %s\n", r.URL.Path, time.Since(start))
	})
}

func main() {
	l := lion.New()
	l.Use(&logger{})
	l.GetFunc("/", Home)
	l.GetFunc("/hello/:name", Hello)
	l.Run(":3000")
}

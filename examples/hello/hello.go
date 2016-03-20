package main

import (
	"fmt"
	"net/http"

	"github.com/celrenheit/lion"
	"golang.org/x/net/context"
)

func Home(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home\n")
}

func Hello(c context.Context, w http.ResponseWriter, r *http.Request) {
	name := lion.Param(c, "name")
	fmt.Fprintf(w, "Hello "+name)
}

func main() {
	l := lion.Classic()
	l.GetFunc("/", Home)
	l.GetFunc("/hello/:name", Hello)
	l.Run()
}

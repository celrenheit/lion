package main

import (
	"fmt"
	"net/http"

	"github.com/celrenheit/lion"
	"golang.org/x/net/context"
)

func home(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home")
}

func hello(c context.Context, w http.ResponseWriter, r *http.Request) {
	name := lion.Param(c, "name")
	fmt.Fprintf(w, "Hello "+name)
}

func main() {
	l := lion.Classic()
	l.GetFunc("/", home)
	l.GetFunc("/hello/:name", hello)
	l.Run()
}

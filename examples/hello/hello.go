package main

import (
	"fmt"
	"net/http"

	"github.com/celrenheit/lion"
	"github.com/celrenheit/lion/middleware"
)

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home")
}

func hello(w http.ResponseWriter, r *http.Request) {
	name := lion.Param(r, "name")
	fmt.Fprintf(w, "Hello "+name)
}

func main() {
	l := lion.New(middleware.Classic())
	l.GetFunc("/", home)
	l.GetFunc("/hello/:name", hello)
	l.Run()
}

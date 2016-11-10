// Package lion is a fast HTTP router for building modern scalable modular REST APIs in Go.
//
// Install and update:
//    go get -u github.com/celrenheit/lion
//
// Getting started:
//
// Start by importing "github.com/celrenheit/lion" into your project.
// Then you need to create a new instance of the router using lion.New() for a blank router or lion.Classic() for a router with default middlewares included.
//
// Here is a simple hello world example:
//
//		 package main
//
//		 import (
//		 	"fmt"
//		 	"net/http"
//
//		 	"github.com/celrenheit/lion"
//		 	"context"
//		 )
//
//		 func Home(w http.ResponseWriter, r *http.Request) {
//		 	fmt.Fprintf(w, "Home")
//		 }
//
//		 func Hello(w http.ResponseWriter, r *http.Request) {
//		 	name := lion.Param(c, "name")
//		 	fmt.Fprintf(w, "Hello "+name)
//		 }
//
//		 func main() {
//		 	l := lion.Classic()
//		 	l.GetFunc("/", Home)
//		 	l.GetFunc("/hello/:name", Hello)
//		 	l.Run()
//		 }
//
// You can open your web browser to http://localhost:3000/hello/world and you should see "Hello world".
// If it finds a PORT environnement variable it will use that. Otherwise, it will use run the server at localhost:3000.
// If you wish to provide a specific port you can run it using: l.Run(":8080")
//
package lion

import "net/http"

// Middleware interface that takes as input a Handler and returns a Handler
type Middleware interface {
	ServeNext(http.Handler) http.Handler
}

// MiddlewareFunc wraps a function that takes as input a Handler and returns a Handler. So that it implements the Middlewares interface
type MiddlewareFunc func(http.Handler) http.Handler

// ServeNext makes MiddlewareFunc implement Middleware
func (m MiddlewareFunc) ServeNext(next http.Handler) http.Handler {
	return m(next)
}

// Middlewares is an array of Middleware
type Middlewares []Middleware

// BuildHandler builds a chain of middlewares from a passed Handler and returns a Handler
func (middlewares Middlewares) BuildHandler(handler http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i].ServeNext(handler)
	}
	return handler
}

func (mws Middlewares) ServeNext(next http.Handler) http.Handler {
	return mws.BuildHandler(next)
}

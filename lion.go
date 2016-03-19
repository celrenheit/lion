package lion

import (
	"net/http"

	"golang.org/x/net/context"
)

// Handler responds to an HTTP request
type Handler interface {
	ServeHTTPC(context.Context, http.ResponseWriter, *http.Request)
}

// HandlerFunc is a wrapper for a function to implement the Handler interface
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTP makes HandlerFunc implement net/http.Handler interface
func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(context.TODO(), w, r)
}

// ServeHTTPC makes HandlerFunc implement Handler interface
func (h HandlerFunc) ServeHTTPC(c context.Context, w http.ResponseWriter, r *http.Request) {
	h(c, w, r)
}

// Middleware interface that takes as input a Handler and returns a Handler
type Middleware interface {
	ServeNext(Handler) Handler
}

// MiddlewareFunc wraps a function that takes as input a Handler and returns a Handler. So that it implements the Middlewares interface
type MiddlewareFunc func(Handler) Handler

// ServeNext makes MiddlewareFunc implement Middleware
func (m MiddlewareFunc) ServeNext(next Handler) Handler {
	return m(next)
}

// Middlewares is an array of Middleware
type Middlewares []Middleware

func (middlewares Middlewares) BuildHandler(handler Handler) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i].ServeNext(handler)
	}
	return handler
}

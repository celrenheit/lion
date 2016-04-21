package htest

import (
	"io"
	"net/http"
	"testing"
)

// HTTPTester is an interface that represents the main entry point of HTest
type HTTPTester interface {
	Get(path string) Requester
	Head(path string) Requester
	Post(path string) Requester
	Put(path string) Requester
	Delete(path string) Requester
	Trace(path string) Requester
	Options(path string) Requester
	Connect(path string) Requester
	Patch(path string) Requester

	// Request allows to create a Requester object with the passed HTTP method and path
	Request(method, path string) Requester

	// Request allows to create a Requester object with the passed HTTP method, path and body.
	RequestWithBody(method, path string, body io.Reader) Requester

	// SetHandler changes the underlying http.Handler
	SetHandler(handler http.Handler) HTTPTester
}

type htest struct {
	handler http.Handler
	t       testing.TB
}

// New returns a new HTTPTester instance
func New(t testing.TB, handler http.Handler) HTTPTester {
	return &htest{
		handler: handler,
		t:       t,
	}
}

func (h *htest) SetHandler(handler http.Handler) HTTPTester {
	h.handler = handler
	return h
}

func (h *htest) Get(path string) Requester {
	return h.Request("GET", path)
}
func (h *htest) Head(path string) Requester {
	return h.Request("HEAD", path)
}
func (h *htest) Post(path string) Requester {
	return h.Request("POST", path)
}
func (h *htest) Put(path string) Requester {
	return h.Request("PUT", path)
}
func (h *htest) Delete(path string) Requester {
	return h.Request("DELETE", path)
}
func (h *htest) Trace(path string) Requester {
	return h.Request("TRACE", path)
}
func (h *htest) Options(path string) Requester {
	return h.Request("OPTIONS", path)
}
func (h *htest) Connect(path string) Requester {
	return h.Request("CONNECT", path)
}
func (h *htest) Patch(path string) Requester {
	return h.Request("PATCH", path)
}

func (h *htest) Request(method, path string) Requester {
	return h.RequestWithBody(method, path, nil)
}

func (h *htest) RequestWithBody(method, path string, body io.Reader) Requester {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		h.t.Error(err)
		h.t.FailNow()
	}
	return &requester{
		method:  method,
		path:    path,
		body:    body,
		t:       h.t,
		handler: h.handler,
		request: req,
	}
}

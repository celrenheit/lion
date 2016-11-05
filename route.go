package lion

import (
	"net/http"

	"github.com/celrenheit/lion/internal/matcher"
)

type Route interface {
	WithName(name string) Route
	// WithMethod adds a new handler to the corresponding HTTP method.
	// The handler will not be built with middlewares.
	// If you want to add middleware you should add them by yourself.
	WithMethod(method string, handler http.Handler) Route

	Methods() (methods []string)
	Host() string
	Name() string
	Pattern() string
	Handler(method string) http.Handler
	Path(pairs ...string) string
}

type route struct {
	host, name, pattern string
	mws                 Middlewares

	get     http.Handler
	head    http.Handler
	post    http.Handler
	put     http.Handler
	delete  http.Handler
	trace   http.Handler
	options http.Handler
	connect http.Handler
	patch   http.Handler
}

func newRoute() *route {
	return &route{}
}

func (r *route) Uses(mws ...Middleware) Route {
	r.mws = append(r.mws, mws...)
	return r
}

func (r *route) WithName(name string) Route {
	r.name = name
	return r
}

func (r *route) WithPattern(pattern string) Route {
	r.pattern = pattern
	return r
}

func (r *route) WithMethod(method string, handler http.Handler) Route {
	r.addHandler(method, handler)
	return r
}

func (r *route) Methods() (methods []string) {
	for _, m := range allowedHTTPMethods {
		if r.getHandler(m) != nil {
			methods = append(methods, m)
		}
	}
	return
}

func (r *route) Host() string {
	return r.host
}

func (r *route) Name() string {
	return r.name
}

func (r *route) Pattern() string {
	return r.pattern
}

func (r *route) Path(pairs ...string) string {
	// TODO: generate path from pairs parameter
	return r.pattern
}

func (r *route) Handler(method string) http.Handler {
	return r.getHandler(method)
}

func (gs *route) Set(value interface{}, tags matcher.Tags) {
	if len(tags) != 1 {
		panicl("Length != 1")
	}

	method := tags[0]

	var handler http.Handler
	if value == nil {
		handler = nil
	} else {
		if h, ok := value.(http.Handler); !ok {
			panicl("Not handler")
		} else {
			handler = h
		}
	}

	gs.addHandler(method, handler)
}

func (gs *route) Get(tags matcher.Tags) interface{} {
	if len(tags) != 1 {
		return nil
	}

	method := tags[0]

	return gs.getHandler(method)
}

func (gs *route) addHandler(method string, handler http.Handler) {
	switch method {
	case GET:
		gs.get = handler
	case HEAD:
		gs.head = handler
	case POST:
		gs.post = handler
	case PUT:
		gs.put = handler
	case DELETE:
		gs.delete = handler
	case TRACE:
		gs.trace = handler
	case OPTIONS:
		gs.options = handler
	case CONNECT:
		gs.connect = handler
	case PATCH:
		gs.patch = handler
	}
}

func (gs *route) getHandler(method string) http.Handler {
	switch method {
	case GET:
		return gs.get
	case HEAD:
		return gs.head
	case POST:
		return gs.post
	case PUT:
		return gs.put
	case DELETE:
		return gs.delete
	case TRACE:
		return gs.trace
	case OPTIONS:
		return gs.options
	case CONNECT:
		return gs.connect
	case PATCH:
		return gs.patch
	default:
		return nil
	}
}

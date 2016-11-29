package lion

import (
	"net/http"
	"strings"

	"github.com/celrenheit/lion/internal/matcher"
)

// TODO: add this later
// WithMethod adds a new handler to the corresponding HTTP method.
// The handler will not be built with middlewares.
// If you want to add middleware you should add them by yourself.
// WithMethod(method string, handler http.Handler) Route

// Routes is an slice of Route.
// Check Routes.ByName or Routes.ByPattern to find out if it is useful to you
type Routes []Route

func (rs Routes) String() string {
	sa := make([]string, 0, len(rs))
	for _, r := range rs {
		sa = append(sa, r.Pattern())
	}
	return strings.Join(sa, ", ")
}

// ByName returns the route corresponding to the name given.
// It returns nil otherwise.
func (rs Routes) ByName(name string) Route {
	// Since all routes have their name empty by default,
	// we cannot return the first route with an empty name
	if name == "" {
		return nil
	}
	for _, route := range rs {
		if route.Name() == name {
			return route
		}
	}

	return nil
}

// ByPattern returns the route corresponding to the pattern given.
// It returns nil otherwise.
func (rs Routes) ByPattern(pattern string) Route {
	if pattern == "" {
		return nil
	}
	for _, route := range rs {
		if route.Pattern() == pattern {
			return route
		}
	}

	return nil
}

// Route defines a single route registered in your Router.
// A Route corresponds to the pattern and host given.
// It contains the handlers for each HTTP methods.
type Route interface {
	WithName(name string) Route

	Methods() (methods []string)
	Host() string
	Name() string
	Pattern() string
	Handler(method string) http.Handler

	// Path building
	Path(params map[string]string) (string, error)
	Build() RoutePathBuilder
	// Convenient alias for Build().WithParam()
	// Calling this method will create a new RoutePathBuilder
	WithParam(key, value string) RoutePathBuilder
}

type route struct {
	host, name, pattern string

	pathMatcher registerMatcher

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

func (r *route) WithName(name string) Route {
	r.name = name
	return r
}

func (r *route) WithPattern(pattern string) Route {
	r.pattern = pattern
	return r
}

func (r *route) withMethods(handler http.Handler, methods ...string) Route {
	for _, method := range methods {
		r.addHandler(method, handler)
	}
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

func (r *route) Path(params map[string]string) (string, error) {
	return r.pathMatcher.Path(r.Pattern(), params)
}

func (r *route) Handler(method string) http.Handler {
	return r.getHandler(method)
}

func (r *route) Set(value interface{}, tags matcher.Tags) {
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

	r.addHandler(method, handler)
}

func (r *route) Get(tags matcher.Tags) interface{} {
	if len(tags) != 1 {
		return nil
	}

	method := tags[0]

	return r.getHandler(method)
}

func (r *route) addHandler(method string, handler http.Handler) {
	switch method {
	case GET:
		r.get = handler
	case HEAD:
		r.head = handler
	case POST:
		r.post = handler
	case PUT:
		r.put = handler
	case DELETE:
		r.delete = handler
	case TRACE:
		r.trace = handler
	case OPTIONS:
		r.options = handler
	case CONNECT:
		r.connect = handler
	case PATCH:
		r.patch = handler
	}
}

func (r *route) getHandler(method string) http.Handler {
	switch method {
	case GET:
		return r.get
	case HEAD:
		return r.head
	case POST:
		return r.post
	case PUT:
		return r.put
	case DELETE:
		return r.delete
	case TRACE:
		return r.trace
	case OPTIONS:
		return r.options
	case CONNECT:
		return r.connect
	case PATCH:
		return r.patch
	default:
		return nil
	}
}

// RoutePathBuilder is a convenient utility to build path given each url parameters.
// Here is a simple example usage.
//		 router := New()
//		 route := router.Get("/posts/:user", postsHandler)
//		 path := route.Build().WithParam("user", "123")
//		 // path should be equal to "/posts/123"
type RoutePathBuilder interface {
	WithParam(key, value string) RoutePathBuilder
	Path() (string, error)
}

type routePathBuilder struct {
	route  *route
	params map[string]string
}

func (r *route) Build() RoutePathBuilder {
	return &routePathBuilder{
		route:  r,
		params: make(map[string]string),
	}
}
func (r *route) WithParam(key, value string) RoutePathBuilder {
	return r.Build().WithParam(key, value)
}

func (r *routePathBuilder) WithParam(key, value string) RoutePathBuilder {
	r.params[key] = value
	return r
}

func (r *routePathBuilder) Path() (string, error) {
	return r.route.Path(r.params)
}

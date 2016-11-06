package lion

import (
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

// HTTP methods constants
const (
	GET     = "GET"
	HEAD    = "HEAD"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	TRACE   = "TRACE"
	OPTIONS = "OPTIONS"
	CONNECT = "CONNECT"
	PATCH   = "PATCH"
)

var allowedHTTPMethods = [...]string{GET, HEAD, POST, PUT, DELETE, TRACE, OPTIONS, CONNECT, PATCH}

// Router is the main component of Lion. It is responsible for registering handlers and middlewares
type Router struct {
	router *Router

	middlewares Middlewares

	handler http.Handler // TODO: create a handler

	pattern string

	notFoundHandler http.Handler

	routes []*route // Used for Mount()

	pool sync.Pool

	namedMiddlewares map[string]Middlewares

	host   string
	hostrm *hostMatcher
}

// New creates a new router instance
func New(mws ...Middleware) *Router {
	r := &Router{
		hostrm:           newHostMatcher(),
		middlewares:      Middlewares{},
		namedMiddlewares: make(map[string]Middlewares),
		pool:             newCtxPool(),
	}
	r.router = r
	r.Use(mws...)
	return r
}

// Subrouter creates a new router based on the parent router.
//
// A subrouter has the same pattern and host as the parent router.
// It has it's own middlewares.
func (r *Router) Subrouter(mws ...Middleware) *Router {
	nr := &Router{
		router:           r,
		hostrm:           r.hostrm,
		pattern:          r.pattern,
		middlewares:      Middlewares{},
		namedMiddlewares: make(map[string]Middlewares),
		host:             r.host,
		pool:             newCtxPool(),
		routes:           []*route{},
	}
	nr.Use(mws...)
	return nr
}

// Group creates a subrouter with parent pattern provided.
func (r *Router) Group(pattern string, mws ...Middleware) *Router {
	p := r.pattern + pattern
	if pattern == "/" && r.pattern != "/" && r.pattern != "" {
		p = r.pattern
	}
	validatePattern(p)

	nr := r.Subrouter(mws...)
	nr.pattern = p
	return nr
}

func newCtxPool() sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			return newContext()
		},
	}
}

// Host sets the host for the current router instances.
// You can use patterns in the same way they are currently used for routes but in reverse order (params on the left)
// 	NOTE: You have to use the '$' character instead of ':' for matching host parameters.
// The following patterns works:
/*
	admin.example.com			will match			admin.example.com
	$username.blog.com			will match			messi.blog.com
						will not match			my.awesome.blog.com
	*.example.com				will match			my.admin.example.com

The following patterns are not allowed:
	mail.*
	*
*/
func (r *Router) Host(hostpattern string) *Router {
	r.host = hostpattern
	r.hostrm.Register(hostpattern)
	return r
}

// Any registers the provided Handler for all of the allowed http methods: GET, HEAD, POST, PUT, DELETE, TRACE, OPTIONS, CONNECT, PATCH
func (r *Router) Any(pattern string, handler http.Handler) Route {
	rt := r.Handle(allowedHTTPMethods[0], pattern, handler).(*route)
	rt.withMethods(r.middlewares.BuildHandler(handler), allowedHTTPMethods[1:]...)
	return rt
}

// Get registers an http GET method receiver with the provided Handler
func (r *Router) Get(pattern string, handler http.Handler) Route {
	return r.Handle("GET", pattern, handler)
}

// Head registers an http HEAD method receiver with the provided Handler
func (r *Router) Head(pattern string, handler http.Handler) Route {
	return r.Handle("HEAD", pattern, handler)
}

// Post registers an http POST method receiver with the provided Handler
func (r *Router) Post(pattern string, handler http.Handler) Route {
	return r.Handle("POST", pattern, handler)
}

// Put registers an http PUT method receiver with the provided Handler
func (r *Router) Put(pattern string, handler http.Handler) Route {
	return r.Handle("PUT", pattern, handler)
}

// Delete registers an http DELETE method receiver with the provided Handler
func (r *Router) Delete(pattern string, handler http.Handler) Route {
	return r.Handle("DELETE", pattern, handler)
}

// Trace registers an http TRACE method receiver with the provided Handler
func (r *Router) Trace(pattern string, handler http.Handler) Route {
	return r.Handle("TRACE", pattern, handler)
}

// Options registers an http OPTIONS method receiver with the provided Handler
func (r *Router) Options(pattern string, handler http.Handler) Route {
	return r.Handle("OPTIONS", pattern, handler)
}

// Connect registers an http CONNECT method receiver with the provided Handler
func (r *Router) Connect(pattern string, handler http.Handler) Route {
	return r.Handle("CONNECT", pattern, handler)
}

// Patch registers an http PATCH method receiver with the provided Handler
func (r *Router) Patch(pattern string, handler http.Handler) Route {
	return r.Handle("PATCH", pattern, handler)
}

// AnyFunc registers the provided HandlerFunc for all of the allowed http methods: GET, HEAD, POST, PUT, DELETE, TRACE, OPTIONS, CONNECT, PATCH
func (r *Router) AnyFunc(pattern string, handler http.HandlerFunc) Route {
	return r.Any(pattern, http.HandlerFunc(handler))
}

// GetFunc wraps a HandlerFunc as a Handler and registers it to the router
func (r *Router) GetFunc(pattern string, fn http.HandlerFunc) Route {
	return r.Get(pattern, http.HandlerFunc(fn))
}

// HeadFunc wraps a HandlerFunc as a Handler and registers it to the router
func (r *Router) HeadFunc(pattern string, fn http.HandlerFunc) Route {
	return r.Head(pattern, http.HandlerFunc(fn))
}

// PostFunc wraps a HandlerFunc as a Handler and registers it to the router
func (r *Router) PostFunc(pattern string, fn http.HandlerFunc) Route {
	return r.Post(pattern, http.HandlerFunc(fn))
}

// PutFunc wraps a HandlerFunc as a Handler and registers it to the router
func (r *Router) PutFunc(pattern string, fn http.HandlerFunc) Route {
	return r.Put(pattern, http.HandlerFunc(fn))
}

// DeleteFunc wraps a HandlerFunc as a Handler and registers it to the router
func (r *Router) DeleteFunc(pattern string, fn http.HandlerFunc) Route {
	return r.Delete(pattern, http.HandlerFunc(fn))
}

// TraceFunc wraps a HandlerFunc as a Handler and registers it to the router
func (r *Router) TraceFunc(pattern string, fn http.HandlerFunc) Route {
	return r.Trace(pattern, http.HandlerFunc(fn))
}

// OptionsFunc wraps a HandlerFunc as a Handler and registers it to the router
func (r *Router) OptionsFunc(pattern string, fn http.HandlerFunc) Route {
	return r.Options(pattern, http.HandlerFunc(fn))
}

// ConnectFunc wraps a HandlerFunc as a Handler and registers it to the router
func (r *Router) ConnectFunc(pattern string, fn http.HandlerFunc) Route {
	return r.Connect(pattern, http.HandlerFunc(fn))
}

// PatchFunc wraps a HandlerFunc as a Handler and registers it to the router
func (r *Router) PatchFunc(pattern string, fn http.HandlerFunc) Route {
	return r.Patch(pattern, http.HandlerFunc(fn))
}

// Use registers middlewares to be used
func (r *Router) Use(middlewares ...Middleware) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// UseFunc wraps a MiddlewareFunc as a Middleware and registers it middlewares to be used
func (r *Router) UseFunc(middlewareFuncs ...MiddlewareFunc) {
	for _, fn := range middlewareFuncs {
		r.Use(MiddlewareFunc(fn))
	}
}

func (r *Router) UseNext(funcs ...func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)) {
	for _, fn := range funcs {
		r.Use(MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fn(w, r, next.ServeHTTP)
			})
		}))
	}
}

// Handle is the underling method responsible for registering a handler for a specific method and pattern.
func (r *Router) Handle(method, pattern string, handler http.Handler) Route {
	var p string
	if !r.isRoot() && pattern == "/" && r.pattern != "" {
		p = r.pattern
	} else {
		p = r.pattern + pattern
	}

	built := r.buildMiddlewares(handler)

	rm := r.router.hostrm.Register(r.host)
	rt := rm.Register(method, p, built)

	// If this route does not exist in this Router instance then add it
	if _, ok := r.findRoute(rt); !ok {
		rt.pattern = p
		rt.host = r.host
		rt.pathMatcher = rm
		r.routes = append(r.routes, rt)
	}
	return rt
}

func (r *Router) findRoute(rt *route) (*route, bool) {
	for _, route := range r.routes {
		if route == rt {
			return route, true
		}
	}

	return nil, false
}

// Mount mounts a subrouter at the provided pattern
func (r *Router) Mount(pattern string, router *Router, mws ...Middleware) {
	host := r.host
	for _, route := range router.routes {
		r.Host(route.Host())
		for _, method := range route.Methods() {
			r.Handle(method, path.Join(pattern, route.Pattern()), route.Handler(method))
		}
	}

	// Restore previous host
	r.host = host
}

func (r *Router) buildMiddlewares(handler http.Handler) http.Handler {
	handler = r.middlewares.BuildHandler(handler)
	if !r.isRoot() {
		handler = r.router.buildMiddlewares(handler)
	}
	return handler
}

func (r *Router) isRoot() bool {
	return r.router == r
}

// HandleFunc wraps a HandlerFunc and pass it to Handle method
func (r *Router) HandleFunc(method, pattern string, fn http.HandlerFunc) Route {
	return r.Handle(method, pattern, http.HandlerFunc(fn))
}

// ServeHTTP finds the handler associated with the request's path.
// If it is not found it calls the NotFound handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := r.pool.Get().(*ctx)
	ctx.parent = req.Context()
	ctx.ResponseWriter = w
	ctx.req = req

	if h := r.router.hostrm.Match(ctx, req); h != nil {
		// We set the context only if there is a match
		req = setParamContext(req, ctx)

		h.ServeHTTP(w, req)
	} else {
		r.notFound(w, req) // r.middlewares.BuildHandler(HandlerFunc(r.NotFound)).ServeHTTPC
	}

	ctx.Reset()
	r.pool.Put(ctx)
}

// NotFound calls NotFoundHandler() if it is set. Otherwise, it calls net/http.NotFound
func (r *Router) notFound(w http.ResponseWriter, req *http.Request) {
	if r.router.notFoundHandler != nil {
		r.router.notFoundHandler.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}

// NotFoundHandler gives the ability to use a specific 404 NOT FOUND handler
func (r *Router) NotFoundHandler(handler http.Handler) {
	r.notFoundHandler = handler
}

// ServeFiles serves files located in root http.FileSystem
//
// This can be used as shown below:
// 	r := New()
// 	r.ServeFiles("/static", http.Dir("static")) // This will serve files in the directory static with /static prefix
func (r *Router) ServeFiles(base string, root http.FileSystem) {
	if strings.ContainsAny(base, ":*") {
		panic("Lion: ServeFiles cannot have url parameters")
	}

	pattern := path.Join(base, "/*")
	fileServer := http.StripPrefix(base, http.FileServer(root))

	r.Get(pattern, fileServer)
	r.Head(pattern, fileServer)
}

// ServeFile serve a specific file located at the passed path
//
// 	l := New()
// 	l.ServeFile("/robots.txt", "path/to/robots.txt")
func (r *Router) ServeFile(base, path string) {
	if strings.ContainsAny(base, ":*") {
		panic("Lion: ServeFile cannot have url parameters")
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	})

	r.Get(base, handler)
	r.Head(base, handler)
}

// Run calls http.ListenAndServe for the current router.
// If no addresses are specified as arguments, it will use the PORT environnement variable if it is defined. Otherwise, it will listen on port 3000 of the localmachine
//
// 	r := New()
// 	r.Run() // will call
// 	r.Run(":8080")
func (r *Router) Run(addr ...string) {
	var a string

	if len(addr) == 0 {
		if p := os.Getenv("PORT"); p != "" {
			a = ":" + p
		} else {
			a = ":3000"
		}
	} else {
		a = addr[0]
	}

	lionLogger.Printf("listening on %s", a)
	lionLogger.Fatal(http.ListenAndServe(a, r))
}

// RunTLS calls http.ListenAndServeTLS for the current router
//
// 	r := New()
// 	r.RunTLS(":3443", "cert.pem", "key.pem")
func (r *Router) RunTLS(addr, certFile, keyFile string) {
	lionLogger.Printf("listening on %s", addr)
	lionLogger.Fatal(http.ListenAndServeTLS(addr, certFile, keyFile, r))
}

// Define registers some middleware using a name for reuse later using UseNamed method.
func (r *Router) Define(name string, mws ...Middleware) {
	r.namedMiddlewares[name] = append(r.namedMiddlewares[name], mws...)
}

// DefineFunc is a convenience wrapper for Define() to use MiddlewareFunc instead of a Middleware instance
func (r *Router) DefineFunc(name string, mws ...MiddlewareFunc) {
	for _, mw := range mws {
		r.Define(name, mw)
	}
}

// UseNamed adds a middleware already defined using Define method.
// If it cannot find it in the current router, it will look for it in the parent router.
func (r *Router) UseNamed(name string) {
	if r.hasNamed(name) { // Find if it this is registered in the current router
		r.Use(r.namedMiddlewares[name]...)
	} else if !r.isRoot() { // Otherwise, look for it in parent router.
		r.router.UseNamed(name)
	} else { // not found
		panic("Unknow named middlewares: " + name)
	}
}

func (r *Router) hasNamed(name string) bool {
	_, exist := r.namedMiddlewares[name]
	return exist
}

func validatePattern(pattern string) {
	if len(pattern) == 0 || pattern[0] != '/' {
		panic("path must start with '/' in path '" + pattern + "'")
	}
}

// Route get the Route associated with the name specified.
func (r *Router) Route(name string) Route {
	// Since all routes have their name empty by default,
	// we cannot return the first route with an empty name
	if name == "" {
		return nil
	}
	for _, route := range r.routes {
		if route.Name() == name {
			return route
		}
	}

	return nil
}

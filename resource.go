package lion

import (
	"net/http"
	"reflect"
	"strings"
)

// Resource defines the minimum required methods
type Resource interface{}

// resourceUses is an interface with the Uses() method which can be used to define global middlewares for the resource.
type resourceUses interface {
	Uses() Middlewares
}

// Resource registers a Resource with the corresponding pattern
func (r *Router) Resource(pattern string, resource Resource) {
	sub := r.Group(pattern)

	if usesRes, ok := resource.(resourceUses); ok {
		if len(usesRes.Uses()) > 0 {
			sub.Use(usesRes.Uses()...)
		}
	}

	for _, m := range allowedHTTPMethods {
		if hfn, ok := isHandlerFuncInResource(m, resource); ok {
			s := sub.Subrouter()
			if mws, ok := isMiddlewareInResource(m, resource); ok {
				s.Use(mws()...)
			}
			s.HandleFunc(m, "/", http.HandlerFunc(hfn))
		}
	}
}

// checks if there is a Name(w http.ResponseWriter, r *http.Request) method available on the Resource r
func isHandlerFuncInResource(m string, r Resource) (func(w http.ResponseWriter, r *http.Request), bool) {
	name := strings.Title(strings.ToLower(m))
	method := reflect.ValueOf(r).MethodByName(name)
	if !method.IsValid() {
		return nil, false
	}

	// Native http.HandlerFunc
	fn, ok := method.Interface().(func(w http.ResponseWriter, r *http.Request))
	if ok {
		return fn, true
	}

	// ... or check for a contextual handler
	cfn, ok := method.Interface().(func(Context))
	if !ok {
		return nil, false
	}
	return wrapContextHandler(cfn), ok
}

// checks if there is a NameMiddlewares() Middlewares method available on the Resource r
func isMiddlewareInResource(m string, r Resource) (func() Middlewares, bool) {
	name := strings.Title(strings.ToLower(m)) + "Middlewares"
	method := reflect.ValueOf(r).MethodByName(name)
	if !method.IsValid() {
		return nil, false
	}

	fn, ok := method.Interface().(func() Middlewares)
	return fn, ok
}

func wrapContextHandler(fn func(Context)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := C(r)
		if c == nil {
			c = newContextWithResReq(r.Context(), w, r)
		}

		fn(c)
	})
}

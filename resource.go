package lion

import (
	"net/http"
	"reflect"
	"strings"

	"golang.org/x/net/context"
)

// Resource defines the minimum required methods
type Resource interface{}

// ResourceUses is an interface with the Uses() method which can be used to define global middlewares for the resource.
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type ResourceUses interface {
	Uses() Middlewares
}

// GetResourceMiddlewares is an interface for defining middlewares used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type GetResourceMiddlewares interface {
	GetMiddlewares() Middlewares
}

// HeadResourceMiddlewares is an interface for defining middlewares used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type HeadResourceMiddlewares interface {
	HeadMiddlewares() Middlewares
}

// PostResourceMiddlewares is an interface for defining middlewares used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type PostResourceMiddlewares interface {
	PostMiddlewares() Middlewares
}

// PutResourceMiddlewares is an interface for defining middlewares used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type PutResourceMiddlewares interface {
	PutMiddlewares() Middlewares
}

// DeleteResourceMiddlewares is an interface for defining middlewares used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type DeleteResourceMiddlewares interface {
	DeleteMiddlewares() Middlewares
}

// TraceResourceMiddlewares is an interface for defining middlewares used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type TraceResourceMiddlewares interface {
	TraceMiddlewares() Middlewares
}

// OptionsResourceMiddlewares is an interface for defining middlewares used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type OptionsResourceMiddlewares interface {
	OptionsMiddlewares() Middlewares
}

// ConnectResourceMiddlewares is an interface for defining middlewares used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type ConnectResourceMiddlewares interface {
	ConnectMiddlewares() Middlewares
}

// PatchResourceMiddlewares is an interface for defining middlewares used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type PatchResourceMiddlewares interface {
	PatchMiddlewares() Middlewares
}

// GetResource is an interface for defining a HandlerFunc used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type GetResource interface {
	Get(c context.Context, w http.ResponseWriter, r *http.Request)
}

// HeadResource is an interface for defining a HandlerFunc used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type HeadResource interface {
	Head(c context.Context, w http.ResponseWriter, r *http.Request)
}

// PostResource is an interface for defining a HandlerFunc used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type PostResource interface {
	Post(c context.Context, w http.ResponseWriter, r *http.Request)
}

// PutResource is an interface for defining a HandlerFunc used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type PutResource interface {
	Put(c context.Context, w http.ResponseWriter, r *http.Request)
}

// DeleteResource is an interface for defining a HandlerFunc used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type DeleteResource interface {
	Delete(c context.Context, w http.ResponseWriter, r *http.Request)
}

// TraceResource is an interface for defining a HandlerFunc used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type TraceResource interface {
	Trace(c context.Context, w http.ResponseWriter, r *http.Request)
}

// OptionsResource is an interface for defining a HandlerFunc used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type OptionsResource interface {
	Options(c context.Context, w http.ResponseWriter, r *http.Request)
}

// ConnectResource is an interface for defining a HandlerFunc used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type ConnectResource interface {
	Connect(c context.Context, w http.ResponseWriter, r *http.Request)
}

// PatchResource is an interface for defining a HandlerFunc used in Resource method
// DEPRECATED: These methods will be removed in v2.
// This should not cause problems since they are already publicly exposed for documentation purpose only.
type PatchResource interface {
	Patch(c context.Context, w http.ResponseWriter, r *http.Request)
}

// Resource registers a Resource with the corresponding pattern
func (r *Router) Resource(pattern string, resource Resource) {
	sub := r.Group(pattern)

	if usesRes, ok := resource.(ResourceUses); ok {
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
			s.HandleFunc(m, "/", hfn)
		}
	}
}

// checks if there is a Name(c context.Context, w http.ResponseWriter, r *http.Request) method available on the Resource r
func isHandlerFuncInResource(m string, r Resource) (func(c context.Context, w http.ResponseWriter, r *http.Request), bool) {
	name := strings.Title(strings.ToLower(m))
	method := reflect.ValueOf(r).MethodByName(name)
	if !method.IsValid() {
		return nil, false
	}

	fn, ok := method.Interface().(func(c context.Context, w http.ResponseWriter, r *http.Request))
	return fn, ok
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

package lion

import (
	"net/http"

	"golang.org/x/net/context"
)

// Resource defines the minimum required methods
type Resource interface{}

// ResourceUses is an interface with the Uses() method which can be used to define global middlewares for the resource.
type ResourceUses interface {
	Uses() Middlewares
}

// GetResourceMiddlewares is an interface for defining middlewares used in Resource method
type GetResourceMiddlewares interface {
	GetMiddlewares() Middlewares
}

// PostResourceMiddlewares is an interface for defining middlewares used in Resource method
type PostResourceMiddlewares interface {
	PostMiddlewares() Middlewares
}

// PutResourceMiddlewares is an interface for defining middlewares used in Resource method
type PutResourceMiddlewares interface {
	PutMiddlewares() Middlewares
}

// DeleteResourceMiddlewares is an interface for defining middlewares used in Resource method
type DeleteResourceMiddlewares interface {
	DeleteMiddlewares() Middlewares
}

// GetResource is an interface for defining a HandlerFunc used in Resource method
type GetResource interface {
	Get(c context.Context, w http.ResponseWriter, r *http.Request)
}

// PostResource is an interface for defining a HandlerFunc used in Resource method
type PostResource interface {
	Post(c context.Context, w http.ResponseWriter, r *http.Request)
}

// PutResource is an interface for defining a HandlerFunc used in Resource method
type PutResource interface {
	Put(c context.Context, w http.ResponseWriter, r *http.Request)
}

// DeleteResource is an interface for defining a HandlerFunc used in Resource method
type DeleteResource interface {
	Delete(c context.Context, w http.ResponseWriter, r *http.Request)
}

// Resource registers a Resource with the corresponding pattern
func (r *Router) Resource(pattern string, resource Resource) {
	sub := r.Group(pattern)

	if usesRes, ok := resource.(ResourceUses); ok {
		if len(usesRes.Uses()) > 0 {
			sub.Use(usesRes.Uses()...)
		}
	}

	if res, ok := resource.(GetResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(GetResourceMiddlewares); ok {
			s.Use(mw.GetMiddlewares()...)
		}
		s.GetFunc("/", res.Get)
	}

	if res, ok := resource.(PostResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(PostResourceMiddlewares); ok {
			s.Use(mw.PostMiddlewares()...)
		}
		s.PostFunc("/", res.Post)
	}

	if res, ok := resource.(PutResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(PutResourceMiddlewares); ok {
			s.Use(mw.PutMiddlewares()...)
		}
		s.PutFunc("/", res.Put)
	}

	if res, ok := resource.(DeleteResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(DeleteResourceMiddlewares); ok {
			s.Use(mw.DeleteMiddlewares()...)
		}
		s.DeleteFunc("/", res.Delete)
	}
}

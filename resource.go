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

// HeadResourceMiddlewares is an interface for defining middlewares used in Resource method
type HeadResourceMiddlewares interface {
	HeadMiddlewares() Middlewares
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

// TraceResourceMiddlewares is an interface for defining middlewares used in Resource method
type TraceResourceMiddlewares interface {
	TraceMiddlewares() Middlewares
}

// OptionsResourceMiddlewares is an interface for defining middlewares used in Resource method
type OptionsResourceMiddlewares interface {
	OptionsMiddlewares() Middlewares
}

// ConnectResourceMiddlewares is an interface for defining middlewares used in Resource method
type ConnectResourceMiddlewares interface {
	ConnectMiddlewares() Middlewares
}

// PatchResourceMiddlewares is an interface for defining middlewares used in Resource method
type PatchResourceMiddlewares interface {
	PatchMiddlewares() Middlewares
}

// GetResource is an interface for defining a HandlerFunc used in Resource method
type GetResource interface {
	Get(c context.Context, w http.ResponseWriter, r *http.Request)
}

// HeadResource is an interface for defining a HandlerFunc used in Resource method
type HeadResource interface {
	Head(c context.Context, w http.ResponseWriter, r *http.Request)
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

// TraceResource is an interface for defining a HandlerFunc used in Resource method
type TraceResource interface {
	Trace(c context.Context, w http.ResponseWriter, r *http.Request)
}

// OptionsResource is an interface for defining a HandlerFunc used in Resource method
type OptionsResource interface {
	Options(c context.Context, w http.ResponseWriter, r *http.Request)
}

// ConnectResource is an interface for defining a HandlerFunc used in Resource method
type ConnectResource interface {
	Connect(c context.Context, w http.ResponseWriter, r *http.Request)
}

// PatchResource is an interface for defining a HandlerFunc used in Resource method
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

	// GET
	if res, ok := resource.(GetResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(GetResourceMiddlewares); ok {
			s.Use(mw.GetMiddlewares()...)
		}
		s.GetFunc("/", res.Get)
	}

	// HEAD
	if res, ok := resource.(HeadResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(HeadResourceMiddlewares); ok {
			s.Use(mw.HeadMiddlewares()...)
		}
		s.HeadFunc("/", res.Head)
	}

	// POST
	if res, ok := resource.(PostResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(PostResourceMiddlewares); ok {
			s.Use(mw.PostMiddlewares()...)
		}
		s.PostFunc("/", res.Post)
	}

	// PUT
	if res, ok := resource.(PutResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(PutResourceMiddlewares); ok {
			s.Use(mw.PutMiddlewares()...)
		}
		s.PutFunc("/", res.Put)
	}

	// DELETE
	if res, ok := resource.(DeleteResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(DeleteResourceMiddlewares); ok {
			s.Use(mw.DeleteMiddlewares()...)
		}
		s.DeleteFunc("/", res.Delete)
	}

	// TRACE
	if res, ok := resource.(TraceResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(TraceResourceMiddlewares); ok {
			s.Use(mw.TraceMiddlewares()...)
		}
		s.TraceFunc("/", res.Trace)
	}

	// OPTIONS
	if res, ok := resource.(OptionsResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(OptionsResourceMiddlewares); ok {
			s.Use(mw.OptionsMiddlewares()...)
		}
		s.OptionsFunc("/", res.Options)
	}

	// CONNECT
	if res, ok := resource.(ConnectResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(ConnectResourceMiddlewares); ok {
			s.Use(mw.ConnectMiddlewares()...)
		}
		s.ConnectFunc("/", res.Connect)
	}

	// PATCH
	if res, ok := resource.(PatchResource); ok {
		s := sub.Group("/")
		if mw, ok := resource.(PatchResourceMiddlewares); ok {
			s.Use(mw.PatchMiddlewares()...)
		}
		s.PatchFunc("/", res.Patch)
	}
}

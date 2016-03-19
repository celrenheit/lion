package lion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"
)

type testResource struct{}

func (tr testResource) Uses() Middlewares { return Middlewares{} }

func (tr testResource) GetMiddlewares() Middlewares {
	return Middlewares{MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("foo", "Get")
			next.ServeHTTPC(c, w, r)
		})
	})}
}

func (tr testResource) PostMiddlewares() Middlewares {
	return Middlewares{MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("foo", "Post")
			next.ServeHTTPC(c, w, r)
		})
	})}
}
func (tr testResource) PutMiddlewares() Middlewares {
	return Middlewares{MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("foo", "Put")
			next.ServeHTTPC(c, w, r)
		})
	})}
}
func (tr testResource) DeleteMiddlewares() Middlewares {
	return Middlewares{MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("foo", "Delete")
			next.ServeHTTPC(c, w, r)
		})
	})}
}

func (tr testResource) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get"))
}
func (tr testResource) Post(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Post"))
}
func (tr testResource) Put(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Put"))
}
func (tr testResource) Delete(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Delete"))
}

func TestResources(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	expected := []string{"Get", "Post", "Put", "Delete"}
	tr := testResource{}
	// hfuncs := []HandlerFunc{tr.Get, tr.Post, tr.Put, tr.Delete}

	r := New()
	r.Resource("/testpath", tr)

	for i := 0; i < len(methods); i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(methods[i], "/testpath", nil)

		r.ServeHTTP(w, req)

		if w.Body.String() != expected[i] {
			t.Errorf("[Resource] Expected body %s but got %s for http method %s", expected[i], w.Body.String(), methods[i])
		}

		if w.Header().Get("foo") != expected[i] {
			t.Errorf("[Resource] Expected header %s but got %s for http method %s", expected[i], w.Header().Get("foo"), methods[i])
		}
	}
}

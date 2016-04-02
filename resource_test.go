package lion

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"
)

type testResource struct{}

func (tr testResource) Uses() Middlewares { return Middlewares{} }

func (tr testResource) GetMiddlewares() Middlewares     { return Middlewares{newTestResMW("Get")} }
func (tr testResource) HeadMiddlewares() Middlewares    { return Middlewares{newTestResMW("Head")} }
func (tr testResource) PostMiddlewares() Middlewares    { return Middlewares{newTestResMW("Post")} }
func (tr testResource) PutMiddlewares() Middlewares     { return Middlewares{newTestResMW("Put")} }
func (tr testResource) DeleteMiddlewares() Middlewares  { return Middlewares{newTestResMW("Delete")} }
func (tr testResource) TraceMiddlewares() Middlewares   { return Middlewares{newTestResMW("Trace")} }
func (tr testResource) OptionsMiddlewares() Middlewares { return Middlewares{newTestResMW("Options")} }
func (tr testResource) ConnectMiddlewares() Middlewares { return Middlewares{newTestResMW("Connect")} }
func (tr testResource) PatchMiddlewares() Middlewares   { return Middlewares{newTestResMW("Patch")} }

func (tr testResource) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Get")
}
func (tr testResource) Head(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Head")
}
func (tr testResource) Post(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Post")
}
func (tr testResource) Put(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Put")
}
func (tr testResource) Delete(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Delete")
}
func (tr testResource) Trace(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Trace")
}
func (tr testResource) Options(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Options")
}
func (tr testResource) Connect(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Connect")
}
func (tr testResource) Patch(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Patch")
}

func TestResources(t *testing.T) {
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "OPTIONS", "CONNECT", "PATCH"}
	expected := []string{"Get", "Head", "Post", "Put", "Delete", "Trace", "Options", "Connect", "Patch"}
	tr := testResource{}

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

func newTestResMW(header string) Middleware {
	return MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("foo", header)
			next.ServeHTTPC(c, w, r)
		})
	})
}

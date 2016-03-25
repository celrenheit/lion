package lion

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
)

var (
	emptyParams = map[string]string{}
)

func TestRouteMatching(t *testing.T) {
	helloHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	helloNameHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	helloNameTweetsHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	helloNameGetTweetHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	cartsHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	getCartHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	helloContactHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	helloContactNamedHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	helloContactNamedDeeperHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	helloContactNamedSubParamHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	helloContactByPersonHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	helloContactByPersonStaticHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	helloContactByPersonToPersonHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	helloContactByPersonAndPathHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	extensionHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	usernameHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	wildcardHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})

	userProfileHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	userSuperHandler := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})
	userMainWildcard := HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {})

	routes := []struct {
		Method  string
		Pattern string
		Handler Handler
	}{
		{
			Method:  "GET",
			Pattern: "/hello",
			Handler: helloHandler,
		},
		{
			Method:  "GET",
			Pattern: "/hello/contact",
			Handler: helloContactHandler,
		},
		{
			Method:  "GET",
			Pattern: "/hello/:name",
			Handler: helloNameHandler,
		},
		{
			Method:  "GET",
			Pattern: "/hello/:name/tweets",
			Handler: helloNameTweetsHandler,
		},
		{
			Method:  "GET",
			Pattern: "/hello/:name/tweets/:id",
			Handler: helloNameGetTweetHandler,
		},
		{
			Method:  "GET",
			Pattern: "/carts",
			Handler: cartsHandler,
		},
		{
			Method:  "GET",
			Pattern: "/carts/:cartid",
			Handler: getCartHandler,
		},
		{
			Method:  "GET",
			Pattern: "/hello/contact/named",
			Handler: helloContactNamedHandler,
		},
		{
			Method:  "GET",
			Pattern: "/hello/contact/named/deeper",
			Handler: helloContactNamedDeeperHandler,
		},
		{
			Method:  "GET",
			Pattern: "/hello/contact/named/:param",
			Handler: helloContactNamedSubParamHandler,
		},
		{
			Method:  "GET",
			Pattern: "/hello/contact/:dest",
			Handler: helloContactByPersonHandler,
		},
		{
			Method:  "GET",
			Pattern: "/hello/contact/:dest/static",
			Handler: helloContactByPersonStaticHandler,
		},
		{
			Method:  "GET",
			Pattern: "/hello/contact/:dest/:from",
			Handler: helloContactByPersonToPersonHandler,
		},
		{
			Method:  "GET",
			Pattern: "/hello/contact/:dest/*path",
			Handler: helloContactByPersonAndPathHandler,
		},
		{
			Method:  "GET",
			Pattern: "/extension/:file.:ext",
			Handler: extensionHandler,
		},
		{
			Method:  "GET",
			Pattern: "/@:username",
			Handler: usernameHandler,
		},
		{
			Method:  "GET",
			Pattern: "/*",
			Handler: wildcardHandler,
		},
		{
			Method:  "GET",
			Pattern: "/users/:userID/profile",
			Handler: userProfileHandler,
		},
		{
			Method:  "GET",
			Pattern: "/users/super/*",
			Handler: userSuperHandler,
		},
		{
			Method:  "GET",
			Pattern: "/users/*",
			Handler: userMainWildcard,
		},
	}

	tests := []struct {
		Input           string
		ExpectedHandler Handler
		ExpectedParams  map[string]string
	}{
		{
			Input:           "/hello",
			ExpectedHandler: helloHandler,
			ExpectedParams:  emptyParams,
		},
		{
			Input:           "/hello/batman",
			ExpectedHandler: helloNameHandler,
			ExpectedParams:  map[string]string{"name": "batman"},
		},
		{
			Input:           "/hello/dot.inthemiddle",
			ExpectedHandler: helloNameHandler,
			ExpectedParams:  map[string]string{"name": "dot.inthemiddle"},
		},
		{
			Input:           "/hello/batman/tweets",
			ExpectedHandler: helloNameTweetsHandler,
			ExpectedParams:  map[string]string{"name": "batman"},
		},
		{
			Input:           "/hello/batman/tweets/123",
			ExpectedHandler: helloNameGetTweetHandler,
			ExpectedParams:  map[string]string{"name": "batman", "id": "123"},
		},
		{
			Input:           "/carts",
			ExpectedHandler: cartsHandler,
			ExpectedParams:  emptyParams,
		},
		{
			Input:           "/carts/123456",
			ExpectedHandler: getCartHandler,
			ExpectedParams:  map[string]string{"cartid": "123456"},
		},
		{
			Input:           "/hello/contact",
			ExpectedHandler: helloContactHandler,
			ExpectedParams:  emptyParams,
		},
		{
			Input:           "/hello/contact/named",
			ExpectedHandler: helloContactNamedHandler,
			ExpectedParams:  emptyParams,
		},
		{
			Input:           "/hello/contact/named/deeper",
			ExpectedHandler: helloContactNamedDeeperHandler,
			ExpectedParams:  emptyParams,
		},
		{
			Input:           "/hello/contact/named/batman",
			ExpectedHandler: helloContactNamedSubParamHandler,
			ExpectedParams:  map[string]string{"param": "batman"},
		},
		{
			Input:           "/hello/contact/nameddd",
			ExpectedHandler: helloContactByPersonHandler,
			ExpectedParams:  map[string]string{"dest": "nameddd"},
		},
		{
			Input:           "/hello/contact/batman",
			ExpectedHandler: helloContactByPersonHandler,
			ExpectedParams:  map[string]string{"dest": "batman"},
		},
		{
			Input:           "/hello/contact/batman/static",
			ExpectedHandler: helloContactByPersonStaticHandler,
			ExpectedParams:  map[string]string{"dest": "batman"},
		},
		{
			Input:           "/hello/contact/batman/robin",
			ExpectedHandler: helloContactByPersonToPersonHandler,
			ExpectedParams:  map[string]string{"dest": "batman", "from": "robin"},
		},
		{
			Input:           "/hello/contact/batman/folder/subfolder/file",
			ExpectedHandler: helloContactByPersonAndPathHandler,
			ExpectedParams:  map[string]string{"dest": "batman", "*": "folder/subfolder/file"},
		},
		{
			Input:           "/extension/batman.jpg",
			ExpectedHandler: extensionHandler,
			ExpectedParams:  map[string]string{"file": "batman", "ext": "jpg"},
		},
		{
			Input:           "/@celrenheit",
			ExpectedHandler: usernameHandler,
			ExpectedParams:  map[string]string{"username": "celrenheit"},
		},
		{
			Input:           "/unkownpath/subfolder",
			ExpectedHandler: wildcardHandler,
			ExpectedParams:  map[string]string{"*": "unkownpath/subfolder"},
		},
		{
			Input:           "/users/123/profile",
			ExpectedHandler: userProfileHandler,
			ExpectedParams:  map[string]string{"userID": "123"},
		},
		{
			Input:           "/users/super/123/okay/yes",
			ExpectedHandler: userSuperHandler,
			ExpectedParams:  map[string]string{"*": "123/okay/yes"},
		},
		{
			Input:           "/users/123/okay/yes",
			ExpectedHandler: userMainWildcard,
			ExpectedParams:  map[string]string{"*": "123/okay/yes"},
		},
	}

	mux := New()
	for _, r := range routes {
		mux.Handle(r.Method, r.Pattern, r.Handler)
	}

	for _, test := range tests {
		req, _ := http.NewRequest("GET", test.Input, nil)

		c, h := mux.rm.Match(&Context{
			parent: context.TODO(),
		}, req)

		// Compare params
		for k, v := range test.ExpectedParams {
			assert.NotNil(t, c.Value(k))
			actual := c.Value(k).(string)
			if actual != v {
				t.Errorf("Expected key %s to equal %s but got %s for url: %s", cyan(k), green(v), red(actual), test.Input)
			}
		}

		// Compare handlers
		if fmt.Sprintf("%v", h) != fmt.Sprintf("%v", test.ExpectedHandler) {
			t.Errorf("Handler not match for %s", test.Input)
		}

		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		// Compare response code
		if w.Code != http.StatusOK {
			t.Errorf("Response should be 200 OK for %s", test.Input)
		}
	}
}

type testmw struct{}

func (m testmw) ServeNext(next Handler) Handler {
	return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Test-Key", "Test-Value")
		next.ServeHTTPC(c, w, r)
	})
}

func TestMiddleware(t *testing.T) {
	mux := New()
	mux.Use(testmw{})
	mux.Get("/hi", HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hi!"))
	}))

	expectHeader(t, mux, "GET", "/hi", "Test-Key", "Test-Value")
}

func TestMiddlewareFunc(t *testing.T) {
	mux := New()
	mux.UseFunc(func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Test-Key", "Test-Value")
			next.ServeHTTPC(c, w, r)
		})
	})
	mux.Get("/hi", HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hi!"))
	}))

	expectHeader(t, mux, "GET", "/hi", "Test-Key", "Test-Value")
}

func TestMiddlewareChain(t *testing.T) {
	mux := New()
	mux.UseFunc(func(next Handler) Handler {
		return nil
	})
}

func TestMountingSubrouter(t *testing.T) {
	mux := New()

	adminrouter := New()
	adminrouter.GetFunc("/:id", func(c context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("admin", "id")
	})

	mux.Mount("/admin", adminrouter)

	expectHeader(t, mux, "GET", "/admin/123", "admin", "id")
}

func TestGroupSubGroup(t *testing.T) {
	s := New()

	admin := s.Group("/admin")
	sub := admin.Group("/")
	sub.UseFunc(func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Test-Key", "Get")
			next.ServeHTTPC(c, w, r)
		})
	})

	sub.GetFunc("/", func(c context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Get"))
	})

	sub2 := admin.Group("/")
	sub2.UseFunc(func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Test-Key", "Put")
			next.ServeHTTPC(c, w, r)
		})
	})

	sub2.PutFunc("/", func(c context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Put"))
	})

	expectHeader(t, s, "GET", "/admin", "Test-Key", "Get")
	expectHeader(t, s, "PUT", "/admin", "Test-Key", "Put")
}

func TestNamedMiddlewares(t *testing.T) {
	l := New()
	l.DefineFunc("admin", func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Test-Key", "admin")
			next.ServeHTTPC(c, w, r)
		})
	})

	l.DefineFunc("public", func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Test-Key", "public")
			next.ServeHTTPC(c, w, r)
		})
	})

	g := l.Group("/admin")
	g.UseNamed("admin")
	g.GetFunc("/test", func(c context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("admintest"))
	})

	p := l.Group("/public")
	p.UseNamed("public")
	p.GetFunc("/test", func(c context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("publictest"))
	})

	expectHeader(t, l, "GET", "/admin/test", "Test-Key", "admin")
	expectHeader(t, l, "GET", "/public/test", "Test-Key", "public")
	expectBody(t, l, "GET", "/admin/test", "admintest")
	expectBody(t, l, "GET", "/public/test", "publictest")
}

func TestEmptyRouter(t *testing.T) {
	l := New()
	expectStatus(t, l, "GET", "/", http.StatusNotFound)
}

func expectStatus(t *testing.T, mux http.Handler, method, path string, status int) {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != status {
		t.Errorf("Expected status code to be %d but got %d for request: %s %s", status, w.Code, method, path)
	}
}

func expectHeader(t *testing.T, mux http.Handler, method, path, k, v string) {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Header().Get(k) != v {
		t.Errorf("Expected header to be %s but got %s for request: %s %s", v, w.Header().Get(k), method, path)
	}
}

func expectBody(t *testing.T, mux http.Handler, method, path, v string) {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Body.String() != v {
		t.Errorf("Expected body to be %s but got %s for request: %s %s", v, w.Body.String(), method, path)
	}
}

package lion

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/celrenheit/htest"

	"golang.org/x/net/context"
)

var (
	emptyParams = M{}
)

func TestRouteMatching(t *testing.T) {
	helloHandler := fakeHandler()
	helloNameHandler := fakeHandler()
	helloNameTweetsHandler := fakeHandler()
	helloNameGetTweetHandler := fakeHandler()
	cartsHandler := fakeHandler()
	getCartHandler := fakeHandler()
	helloContactHandler := fakeHandler()
	helloContactNamedHandler := fakeHandler()
	helloContactNamedDeeperHandler := fakeHandler()
	helloContactNamedSubParamHandler := fakeHandler()
	helloContactByPersonHandler := fakeHandler()
	helloContactByPersonStaticHandler := fakeHandler()
	helloContactByPersonToPersonHandler := fakeHandler()
	helloContactByPersonAndPathHandler := fakeHandler()
	extensionHandler := fakeHandler()
	usernameHandler := fakeHandler()
	wildcardHandler := fakeHandler()
	userProfileHandler := fakeHandler()
	userSuperHandler := fakeHandler()
	userMainWildcard := fakeHandler()
	emptywildcardHandler := fakeHandler()
	unicodeAlphaHandler := fakeHandler()

	routes := []struct {
		Method  string
		Pattern string
		Handler Handler
	}{
		{Pattern: "/hello", Handler: helloHandler},
		{Pattern: "/hello/contact", Handler: helloContactHandler},
		{Pattern: "/hello/:name", Handler: helloNameHandler},
		{Pattern: "/hello/:name/tweets", Handler: helloNameTweetsHandler},
		{Pattern: "/hello/:name/tweets/:id", Handler: helloNameGetTweetHandler},
		{Pattern: "/carts", Handler: cartsHandler},
		{Pattern: "/carts/:cartid", Handler: getCartHandler},
		{Pattern: "/hello/contact/named", Handler: helloContactNamedHandler},
		{Pattern: "/hello/contact/named/deeper", Handler: helloContactNamedDeeperHandler},
		{Pattern: "/hello/contact/named/:param", Handler: helloContactNamedSubParamHandler},
		{Pattern: "/hello/contact/:dest", Handler: helloContactByPersonHandler},
		{Pattern: "/hello/contact/:dest/static", Handler: helloContactByPersonStaticHandler},
		{Pattern: "/hello/contact/:dest/:from", Handler: helloContactByPersonToPersonHandler},
		{Pattern: "/hello/contact/:dest/*path", Handler: helloContactByPersonAndPathHandler},
		{Pattern: "/extension/:file.:ext", Handler: extensionHandler},
		{Pattern: "/@:username", Handler: usernameHandler},
		{Pattern: "/static/*", Handler: wildcardHandler},
		{Pattern: "/users/:userID/profile", Handler: userProfileHandler},
		{Pattern: "/users/super/*", Handler: userSuperHandler},
		{Pattern: "/users/*", Handler: userMainWildcard},
		{Pattern: "/empty/*", Handler: emptywildcardHandler},
		{Pattern: "/α", Handler: unicodeAlphaHandler},
	}

	tests := []struct {
		Method          string
		Input           string
		ExpectedHandler Handler
		ExpectedParams  M
		ExpectedStatus  int
	}{
		{Input: "/hello", ExpectedHandler: helloHandler, ExpectedParams: emptyParams},
		{Input: "/hello/batman", ExpectedHandler: helloNameHandler, ExpectedParams: M{"name": "batman"}},
		{Input: "/hello/dot.inthemiddle", ExpectedHandler: helloNameHandler, ExpectedParams: M{"name": "dot.inthemiddle"}},
		{Input: "/hello/batman/tweets", ExpectedHandler: helloNameTweetsHandler, ExpectedParams: M{"name": "batman"}},
		{Input: "/hello/batman/tweets/123", ExpectedHandler: helloNameGetTweetHandler, ExpectedParams: M{"name": "batman", "id": "123"}},
		{Input: "/carts", ExpectedHandler: cartsHandler, ExpectedParams: emptyParams},
		{Input: "/carts/123456", ExpectedHandler: getCartHandler, ExpectedParams: M{"cartid": "123456"}},
		{Input: "/hello/contact", ExpectedHandler: helloContactHandler, ExpectedParams: emptyParams},
		{Input: "/hello/contact/named", ExpectedHandler: helloContactNamedHandler, ExpectedParams: emptyParams},
		{Input: "/hello/contact/named/deeper", ExpectedHandler: helloContactNamedDeeperHandler, ExpectedParams: emptyParams},
		{Input: "/hello/contact/named/batman", ExpectedHandler: helloContactNamedSubParamHandler, ExpectedParams: M{"param": "batman"}},
		{Input: "/hello/contact/nameddd", ExpectedHandler: helloContactByPersonHandler, ExpectedParams: M{"dest": "nameddd"}},
		{Input: "/hello/contact/batman", ExpectedHandler: helloContactByPersonHandler, ExpectedParams: M{"dest": "batman"}},
		{Input: "/hello/contact/batman/static", ExpectedHandler: helloContactByPersonStaticHandler, ExpectedParams: M{"dest": "batman"}},
		{Input: "/hello/contact/batman/robin", ExpectedHandler: helloContactByPersonToPersonHandler, ExpectedParams: M{"dest": "batman", "from": "robin"}},
		{Input: "/hello/contact/batman/folder/subfolder/file", ExpectedHandler: helloContactByPersonAndPathHandler, ExpectedParams: M{"dest": "batman", "path": "folder/subfolder/file"}},
		{Input: "/extension/batman.jpg", ExpectedHandler: extensionHandler, ExpectedParams: M{"file": "batman", "ext": "jpg"}},
		{Input: "/@celrenheit", ExpectedHandler: usernameHandler, ExpectedParams: M{"username": "celrenheit"}},
		{Input: "/static/unkownpath/subfolder", ExpectedHandler: wildcardHandler, ExpectedParams: M{"*": "unkownpath/subfolder"}},
		{Input: "/users/123/profile", ExpectedHandler: userProfileHandler, ExpectedParams: M{"userID": "123"}},
		{Input: "/users/super/123/okay/yes", ExpectedHandler: userSuperHandler, ExpectedParams: M{"*": "123/okay/yes"}},
		{Input: "/users/123/okay/yes", ExpectedHandler: userMainWildcard, ExpectedParams: M{"*": "123/okay/yes"}},
		{Input: "/empty/", ExpectedHandler: emptywildcardHandler, ExpectedParams: M{"*": ""}},
		{Input: "/carts404", ExpectedHandler: nil, ExpectedParams: emptyParams, ExpectedStatus: http.StatusNotFound},
		{Input: "/α", ExpectedHandler: unicodeAlphaHandler, ExpectedParams: emptyParams},
		{Input: "/hello/أسد", ExpectedHandler: helloNameHandler, ExpectedParams: M{"name": "أسد"}},
		{Input: "/hello/أسد/tweets", ExpectedHandler: helloNameTweetsHandler, ExpectedParams: M{"name": "أسد"}},
	}

	mux := New()
	for _, r := range routes {
		method := "GET"
		if r.Method != "" {
			method = r.Method
		}
		mux.Handle(method, r.Pattern, r.Handler)
	}

	tester := htest.New(t, mux)

	for _, test := range tests {
		method := "GET"
		if test.Method != "" {
			method = test.Method
		}

		req, _ := http.NewRequest(method, test.Input, nil)

		c, h := mux.rm.Match(&Context{
			parent: context.TODO(),
		}, req)

		// Compare params
		for k, v := range test.ExpectedParams {
			actual := Param(c, k)
			if actual != v {
				t.Errorf("Expected key %s to equal %s but got %s for url: %s", cyan(k), green(v), red(actual), test.Input)
			}
		}

		// Compare handlers
		if fmt.Sprintf("%v", h) != fmt.Sprintf("%v", test.ExpectedHandler) {
			t.Errorf("Handler not match for %s", test.Input)
		}

		expectedStatus := http.StatusOK
		if test.ExpectedStatus != 0 {
			expectedStatus = test.ExpectedStatus
		}

		tester.Request(method, test.Input).Do().
			ExpectStatus(expectedStatus)
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
	htest.New(t, mux).Get("/hi").Do().ExpectHeader("Test-Key", "Test-Value")
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

	htest.New(t, mux).Get("/hi").Do().ExpectHeader("Test-Key", "Test-Value")
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

	htest.New(t, mux).Get("/admin/123").Do().ExpectHeader("admin", "id")
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

	test := htest.New(t, s)
	test.Get("/admin").Do().ExpectHeader("Test-Key", "Get")
	test.Put("/admin").Do().ExpectHeader("Test-Key", "Put")
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

	test := htest.New(t, l)
	test.Get("/admin/test").Do().
		ExpectHeader("Test-Key", "admin").
		ExpectBody("admintest")

	test.Get("/public/test").Do().
		ExpectHeader("Test-Key", "public").
		ExpectBody("publictest")
}

func TestEmptyRouter(t *testing.T) {
	l := New()
	htest.New(t, l).Get("/").Do().ExpectStatus(http.StatusNotFound)
}

func TestServeFiles(t *testing.T) {
	cwd, _ := os.Getwd()
	// Temporary directory
	dir, err := ioutil.TempDir(cwd, "test_serve")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	// Temporary file in the Temporary directory created previously
	f, err := ioutil.TempFile(dir, "")
	f.WriteString("Lion")
	f.Close()

	_, filename := filepath.Split(f.Name())

	l := New()
	test := htest.New(t, l)

	// ServeFiles
	l.ServeFiles("/public", http.Dir(dir))

	// Tests
	test.Get("/public/" + filename).Do().
		ExpectBody("Lion").
		ExpectStatus(http.StatusOK)
	test.Head("/public/"+filename).Do().
		ExpectHeader("Content-type", "text/plain; charset=utf-8").
		ExpectStatus(http.StatusOK)

	// ServeFile
	l.ServeFile("/file", f.Name())

	// Tests
	test.Get("/file").Do().
		ExpectBody("Lion").
		ExpectStatus(http.StatusOK)
	test.Head("/file").Do().
		ExpectHeader("Content-type", "text/plain; charset=utf-8").
		ExpectStatus(http.StatusOK)
}

func TestRouterShouldPanic(t *testing.T) {
	l := New()
	recv := catchPanic(func() {
		l.Get("path", fakeHandler())
	})

	if recv == nil {
		t.Error("Should panic when path does not start with '/'")
	}

	recv = catchPanic(func() {
		l.UseNamed("unknow middleware")
	})

	if recv == nil {
		t.Error("Should panic when using an unknown named middleware")
	}
}

func TestStaticAndWildcardTriggersPanic(t *testing.T) {
	l := New()
	l.Get("/api", fakeHandler())
	l.Get("/*wild", fakeHandler())
	recv := catchPanic(func() {
		htest.New(t, l).Get("/").Do().ExpectStatus(http.StatusOK)
	})
	if recv != nil {
		t.Errorf("Panic triggers")
	}
}

func catchPanic(fn func()) (recv interface{}) {
	defer func() {
		if r := recover(); r != nil {
			recv = r
		}
	}()
	fn()
	return
}

type fakeHandlerType struct{}

func (f *fakeHandlerType) ServeHTTPC(c context.Context, w http.ResponseWriter, r *http.Request) {}

func fakeHandler() Handler {
	return new(fakeHandlerType)
}

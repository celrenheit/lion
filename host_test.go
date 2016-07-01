package lion

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/celrenheit/htest"
)

func TestHostMatcher(t *testing.T) {
	hm := newHostMatcher()

	staticH := fakeHandler()
	demoH := fakeHandler()
	wildH := fakeHandler()
	tldPortH := fakeHandler()
	localhostH := fakeHandler()
	staticPortH := fakeHandler()
	portH := fakeHandler()
	wildSubPortH := fakeHandler()
	ipv4H := fakeHandler()
	ipv4ParamH := fakeHandler()
	ipv4WildH := fakeHandler()
	ipv4WildParamH := fakeHandler()

	toRegister := []struct {
		pattern string
		handler Handler
	}{
		{pattern: "test.batman.com", handler: staticH},
		{pattern: "$demo.batman.com", handler: demoH},
		{pattern: "*.batman.com", handler: wildH},
		{pattern: "batman.$tld:$port", handler: tldPortH},

		{pattern: "localhost", handler: localhostH},
		{pattern: "localhost:1234", handler: staticPortH},
		{pattern: "localhost:$port", handler: portH},
		{pattern: "*.localhost:$port", handler: wildSubPortH},

		{pattern: "01.02.03.04", handler: ipv4H},
		{pattern: "01.$second.03.04", handler: ipv4ParamH},
		{pattern: "*.03.04", handler: ipv4WildH},
		{pattern: "*.03.04:$port", handler: ipv4WildParamH},
	}

	for _, register := range toRegister {
		rm := hm.Register(register.pattern)
		rm.Register(GET, "/", register.handler)
	}

	tests := []struct {
		input           string
		expectedParams  M
		expectedHandler Handler
	}{
		{
			input: "test.batman.com", expectedParams: M{},
			expectedHandler: staticH,
		},
		{
			input: "admin.batman.com", expectedParams: M{"demo": "admin"},
			expectedHandler: demoH,
		},
		{
			input: "forever.batman.com", expectedParams: M{"demo": "forever"},
			expectedHandler: demoH,
		},
		{
			input: "this.is.admin.batman.com", expectedParams: M{"*": "this.is.admin"},
			expectedHandler: wildH,
		},
		{
			input: "batman.org:8080", expectedParams: M{"tld": "org", "port": "8080"},
			expectedHandler: tldPortH,
		},

		// Port
		{
			input: "localhost", expectedParams: emptyParams,
			expectedHandler: localhostH,
		},
		{
			input: "localhost:1234", expectedParams: emptyParams,
			expectedHandler: staticPortH,
		},
		{
			input: "localhost:3000", expectedParams: M{"port": "3000"},
			expectedHandler: portH,
		},
		{
			input: "test.sub.localhost:8080", expectedParams: M{"*": "test.sub", "port": "8080"},
			expectedHandler: wildSubPortH,
		},

		// Ipv4
		{
			input: "01.02.03.04", expectedParams: emptyParams,
			expectedHandler: ipv4H,
		},
		{
			input: "01.99.03.04", expectedParams: M{"second": "99"},
			expectedHandler: ipv4ParamH,
		},
		{
			input: "192.168.03.04", expectedParams: M{"*": "192.168"},
			expectedHandler: ipv4WildH,
		},
		{
			input: "192.168.03.04:3000", expectedParams: M{"*": "192.168", "port": "3000"},
			expectedHandler: ipv4WildParamH,
		},
	}

	for _, test := range tests {
		req, _ := http.NewRequest("GET", "http://"+test.input, nil)
		c := NewContext()
		h := hm.Match(c, req)

		if len(test.expectedParams) != len(c.keys) {
			t.Errorf("Length missmatch: expected %d but got %d (%v) for '%s'", len(test.expectedParams), len(c.keys), c.toMap(), test.input)
		}

		for k, v := range test.expectedParams {
			actual := Param(c, k)
			if actual != v {
				t.Errorf("Expected key %s to equal %s but got %s for host: %s", cyan(k), green(v), red(actual), test.input)
			}
		}

		// Compare handlers
		if fmt.Sprintf("%v", h) != fmt.Sprintf("%v", test.expectedHandler) {
			t.Errorf("Handler not match for %s: expected %v but got %v", test.input, fmt.Sprintf("%v", test.expectedHandler), fmt.Sprintf("%v", h))
		}
	}
}

func TestBasicGroupHost(t *testing.T) {
	mux := New()
	test := htest.New(t, mux)
	mux.Get("/global", fakeHandler())

	mux.Host("*.blog.com")
	mux.Get("/posts", fakeHandler())

	mux.Host("admin.example.com")
	mux.Get("/", fakeHandler())

	group := mux.Group("/users")
	group.Get("/list", fakeHandler())
	group.Host("")
	group.Get("/add", fakeHandler())

	test.Get("/global").Do().ExpectStatus(http.StatusOK)
	test.Get("/").Do().ExpectStatus(http.StatusNotFound)
	test.Get("http://admin.example.com/").Do().ExpectStatus(http.StatusOK)

	test.Get("http://admin.example.com/users/list").Do().ExpectStatus(http.StatusOK)
	test.Get("/users/list").Do().ExpectStatus(http.StatusNotFound)

	test.Get("http://admin.example.com/users/add").Do().ExpectStatus(http.StatusNotFound)
	test.Get("/users/add").Do().ExpectStatus(http.StatusOK)

	test.Get("http://my.awesome.blog.com/posts").Do().ExpectStatus(http.StatusOK)
}

func TestMountHost(t *testing.T) {
	mux := New()
	mux.Host("host1.com")
	mux.Get("/first", fakeHandler())

	second := New()
	second.Host("host2.com")
	second.Get("/second", fakeHandler())

	mux.Mount("/", second)

	test := htest.New(t, mux)
	test.Get("http://host1.com/first").Do().ExpectStatus(http.StatusOK)
	test.Get("http://host1.com/second").Do().ExpectStatus(http.StatusNotFound)
	test.Get("http://host2.com/first").Do().ExpectStatus(http.StatusNotFound)
	test.Get("http://host2.com/second").Do().ExpectStatus(http.StatusOK)
}

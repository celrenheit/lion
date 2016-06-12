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

	toRegister := []struct {
		pattern string
		handler Handler
	}{
		{pattern: "test.batman.com", handler: staticH},
		{pattern: ":demo.batman.com", handler: demoH},
		{pattern: "*.batman.com", handler: wildH},
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
			input: "this.is.admin.batman.com", expectedParams: M{"*": "this.is.admin"},
			expectedHandler: wildH,
		},
	}

	for _, test := range tests {
		req, _ := http.NewRequest("GET", "http://"+test.input, nil)
		c := NewContext()
		h := hm.Match(c, req)

		if len(test.expectedParams) != len(c.keys) {
			t.Errorf("Length missmatch: expected %d but got %d (%v)", len(test.expectedParams), len(c.keys), c.values)
		}

		for k, v := range test.expectedParams {
			actual := Param(c, k)
			if actual != v {
				t.Errorf("Expected key %s to equal %s but got %s for host: %s", cyan(k), green(v), red(actual), test.input)
			}
		}

		// Compare handlers
		if fmt.Sprintf("%v", h) != fmt.Sprintf("%v", test.expectedHandler) {
			t.Errorf("Handler not match for %s: expected %v but got %v", test.input, fmt.Sprintf("%v", h), fmt.Sprintf("%v", test.expectedHandler))
		}
	}
}

func TestBasicGroupHost(t *testing.T) {
	mux := New()
	test := htest.New(t, mux)
	mux.Get("/global", fakeHandler())

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
}

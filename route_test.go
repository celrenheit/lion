package lion

import (
	"testing"

	"github.com/celrenheit/htest"
)

func TestRoute(t *testing.T) {
	l := New()
	l.Get("/hello", fakeHandler()).
		WithMethod(PUT, fakeHandler())

	test := htest.New(t, l)
	test.Get("/hello").Do().ExpectStatus(200)
	test.Put("/hello").Do().ExpectStatus(200)
}

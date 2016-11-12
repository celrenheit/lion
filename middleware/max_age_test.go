package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/celrenheit/htest"
)

func TestMaxAge(t *testing.T) {
	ma := &MaxAge{
		Duration: 3 * time.Hour,
		Filter: func(r *http.Request) bool {
			if r.URL.Path == "/foo" {
				return false
			}
			return true
		},
	}

	handler := ma.ServeNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))

	test := htest.New(t, handler)
	test.Get("/foo").Do().
		ExpectStatus(401).
		ExpectHeader("Cache-Control", "")

	test.Get("/bar").Do().
		ExpectStatus(401).
		ExpectHeader("Cache-Control", "max-age=10800, public, must-revalidate, proxy-revalidate")
}

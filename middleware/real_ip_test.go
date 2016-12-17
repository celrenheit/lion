package middleware

import (
	"net/http"
	"testing"

	"github.com/celrenheit/htest"
)

func TestRealIP(t *testing.T) {
	rip := RealIP{}

	var remoteaddr string
	handler := rip.ServeNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteaddr = r.RemoteAddr
	}))

	test := htest.New(t, handler)
	test.Get("/foo").SetHeader("X-Forwarded-For", "1.2.3.4").Do()
	if remoteaddr != "1.2.3.4" {
		t.Errorf("should equal 1.2.3.4")
	}

	test.Get("/foo").SetHeader("X-Real-IP", "5.6.7.8").Do()
	if remoteaddr != "5.6.7.8" {
		t.Errorf("should equal 5.6.7.8")
	}

}

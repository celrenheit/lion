package middleware

import (
	"net/http"
	"testing"

	"github.com/celrenheit/htest"
)

func TestRequestID(t *testing.T) {
	rid := NewRequestID()

	rid.SetHeader = false

	handler := rid.ServeNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	test := htest.New(t, handler)
	test.Get("/").Do().
		ExpectHeader(headerXRequestID, "")

	rid.SetHeader = true

	handler = rid.ServeNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	test.SetHandler(handler)

	rec := test.Get("/").Do().Recorder()
	if header := rec.Header().Get(headerXRequestID); header == "" {
		t.Errorf("Should have the header %s set", headerXRequestID)
	}
}

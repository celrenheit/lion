package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/celrenheit/htest"
	"github.com/celrenheit/lion"
)

func TestLogger(t *testing.T) {
	r := lion.New()

	buf := new(bytes.Buffer)
	logger := &Logger{
		Logger: log.New(buf, "[lion]", log.Ldate|log.Ltime),
	}

	r.Use(logger)
	r.GetFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	htest.New(t, r).Get("/log").Do().ExpectStatus(200)

	if !strings.Contains(buf.String(), "GET") {
		t.Errorf("Should contain http method GET")
	}

	if !strings.Contains(buf.String(), "/log") {
		t.Errorf("Should contain http path /log")
	}
}

func BenchmarkLogger(b *testing.B) {
	buf := new(bytes.Buffer)
	r := lion.New()
	logger := &Logger{
		Logger: log.New(buf, "[lion]", log.Ldate|log.Ltime),
	}
	r.Use(logger)

	r.GetFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	req, _ := http.NewRequest("GET", "http://localhost/log", nil)
	w := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}

package middleware

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/celrenheit/htest"
)

func TestRecovery(t *testing.T) {
	buf := new(bytes.Buffer)

	recovery := &Recovery{
		Logger:     log.New(buf, "[lion]", log.Ldate|log.Ltime),
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	}

	handler := recovery.ServeNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("OHOH")
	}))

	test := htest.New(t, handler)
	test.Get("/foo").Do().
		ExpectHeader("Content-type", "text/plain; charset=utf-8").
		ExpectBody("").
		ExpectStatus(500)

	if !strings.Contains(buf.String(), "PANIC: OHOH") {
		t.Errorf("Should contain PANIC in log")
	}

	recovery.PrintStack = true
	handler = recovery.ServeNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("OHOH")
	}))

	test = htest.New(t, handler)
	test.Get("/foo").Do().
		ExpectHeader("Content-type", "text/plain; charset=utf-8").
		ExpectBodyContains("PANIC: OHOH").
		ExpectStatus(500)

	if !strings.Contains(buf.String(), "PANIC: OHOH") {
		t.Errorf("Should contain PANIC in log")
	}
}

package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/celrenheit/lion"
)

// Recovery is a middleware that recovers from panics
// Taken from https://github.com/codegangsta/negroni/blob/master/recovery.go
type Recovery struct {
	Logger     *log.Logger
	PrintStack bool
	StackAll   bool
	StackSize  int
}

// NewRecovery creates a new Recovery instance
func NewRecovery() lion.Middleware {
	return &Recovery{
		Logger:     lionLogger,
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	}
}

// ServeNext is the method responsible for recovering from a panic
func (rec *Recovery) ServeNext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if w.Header().Get("Content-type") == "" {
					w.Header().Set("Content-type", "text/plain; charset=utf-8")
				}
				w.WriteHeader(http.StatusInternalServerError)
				stack := make([]byte, rec.StackSize)
				stack = stack[:runtime.Stack(stack, rec.StackAll)]

				f := "PANIC: %s\n%s"
				rec.Logger.Printf(f, err, stack)

				if rec.PrintStack {
					fmt.Fprintf(w, f, err, stack)
				}

			}
		}()

		next.ServeHTTP(w, r)
	})
}

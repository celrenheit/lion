package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
)

// Recovery is a middleware that recovers from panics
// Taken from https://github.com/codegangsta/negroni/blob/master/recovery.go
type Recovery struct {
	Logger     *log.Logger
	PrintStack bool
	StackAll   bool
	StackSize  int
}

// ServeNext is the method responsible for recovering from a panic
func (rec *Recovery) ServeNext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
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

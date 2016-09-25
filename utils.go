package lion

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"golang.org/x/net/context"
)

// Wrap converts an http.Handler to returns a Handler
func Wrap(h http.Handler) Handler {
	return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// WrapFunc converts an http.HandlerFunc to return a Handler
func WrapFunc(fn http.HandlerFunc) Handler {
	return Wrap(http.HandlerFunc(fn))
}

// UnWrap converts a Handler to an http.Handler
func UnWrap(h Handler) http.Handler {
	return HandlerFunc(h.ServeHTTPC)
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	newpath := path.Clean(p)
	if newpath != "/" && p[len(p)-1] == '/' {
		return newpath + "/"
	}
	return newpath
}

func panicl(format string, args ...interface{}) {
	panic(fmt.Sprintf("lion: "+format, args...))
}

func reverseHostStdLib(pattern string) string {
	reversed := strings.Split(pattern, ".")
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return strings.Join(reversed, ".")
}

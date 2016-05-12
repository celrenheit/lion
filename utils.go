package lion

import (
	"fmt"
	"net/http"
	"path"

	"golang.org/x/net/context"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestPrefix(s1, s2 string) int {
	max := min(len(s1), len(s2))
	i := 0
	for i < max && s1[i] == s2[i] {
		i++
	}
	return i
}

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

func stringsIndexAny(str, chars string) int {
	ls := len(str)
	lc := len(chars)

	for i := 0; i < ls; i++ {
		s := str[i]
		for j := 0; j < lc; j++ {
			if s == chars[j] {
				return i
			}
		}
	}
	return -1
}

func stringsIndex(str string, char byte) int {
	ls := len(str)

	for i := 0; i < ls; i++ {
		if str[i] == char {
			return i
		}
	}
	return -1
}

func stringsHasPrefix(str, prefix string) bool {
	// ls := len(str)
	sl := len(str)
	pl := len(prefix)
	if sl < pl {
		return false
	}
	i := 0
	for ; i < pl; i++ {
		if str[i] != prefix[i] {
			break
		}
	}
	if i == pl {
		return true
	}
	return false
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

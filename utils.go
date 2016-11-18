package lion

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

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

func wrap(ctxHandler func(Context)) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c := C(r)
		ctxHandler(c)
	}
	return http.HandlerFunc(fn)
}

package lion

import (
	"net/http"
	"testing"

	"github.com/celrenheit/htest"
)

type testmodule struct {
	base string
}

func (m testmodule) Routes(r *Router) {

}

func (m testmodule) Base() string {
	return m.base
}

func (m testmodule) Requires() []string {
	return []string{"auth", "jwt"}
}

func (m testmodule) Uses() (mws Middlewares) {
	return mws
}

func (m testmodule) Get(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("getmodule"))
}

func TestModule(t *testing.T) {
	l := New()
	l.DefineFunc("auth", func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("auth", "authmw")
			next.ServeHTTP(w, r)
		})
	})

	l.DefineFunc("jwt", func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("token", "jwtmw")
			next.ServeHTTP(w, r)
		})
	})

	l.Module(testmodule{"/admin"})

	test := htest.New(t, l)
	test.Get("/admin").Do().
		ExpectHeader("auth", "authmw").
		ExpectHeader("token", "jwtmw").
		ExpectBody("getmodule")
}

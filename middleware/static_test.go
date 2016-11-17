package middleware

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/celrenheit/htest"
)

func TestStatic(t *testing.T) {

	cwd, _ := os.Getwd()
	// Temporary directory
	dir, err := ioutil.TempDir(cwd, "test_static")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	// Temporary file in the Temporary directory created previously
	f, err := ioutil.TempFile(dir, "")
	if err != nil {
		t.Error(err)
	}
	f.WriteString("Static File")
	f.Close()

	indexf, err := os.Create(dir + "/index.html")
	if err != nil {
		t.Error(err)
	}
	indexf.WriteString("Static Index")
	indexf.Close()

	_, filename := filepath.Split(f.Name())

	static := NewStatic(http.Dir(dir))
	handler := static.ServeNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))

	test := htest.New(t, handler)

	test.Get("/").Do().ExpectBody("Static Index")
	test.Post("/").Do().ExpectStatus(404)
	test.Get("/" + filename).Do().ExpectBody("Static File")
	test.Get("/hello").Do().ExpectStatus(404)
}

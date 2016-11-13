package middleware

import (
	"net/http"
	"path"
	"strings"

	"github.com/celrenheit/lion"
)

// Static is a middleware handler that serves static files in the given directory/filesystem.
// Taken from https://github.com/codegangsta/negroni/blob/master/static.go
type Static struct {
	// Dir is the directory to serve static files from
	Dir http.FileSystem
	// Prefix is the optional prefix used to serve the static directory content
	Prefix string
	// IndexFile defines which file to serve as index if it exists.
	IndexFile string
}

// NewStatic returns a new instance of Static
func NewStatic(directory http.FileSystem) lion.Middleware {
	return &Static{
		Dir:       directory,
		Prefix:    "",
		IndexFile: "index.html",
	}
}

// ServeNext tries to find a file in the directory
func (s *Static) ServeNext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "GET" && r.Method != "HEAD" {
			next.ServeHTTP(w, r)
			return
		}

		file := r.URL.Path
		// if we have a prefix, filter requests by stripping the prefix
		if s.Prefix != "" {
			if !strings.HasPrefix(file, s.Prefix) {
				next.ServeHTTP(w, r)
				return
			}
			file = file[len(s.Prefix):]
			if file != "" && file[0] != '/' {
				next.ServeHTTP(w, r)
				return
			}
		}
		f, err := s.Dir.Open(file)
		if err != nil {
			// discard the error?
			next.ServeHTTP(w, r)
			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// try to serve index file
		if fi.IsDir() {
			// redirect if missing trailing slash
			if !strings.HasSuffix(r.URL.Path, "/") {
				http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
				return
			}

			file = path.Join(file, s.IndexFile)
			f, err = s.Dir.Open(file)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			defer f.Close()

			fi, err = f.Stat()
			if err != nil || fi.IsDir() {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.ServeContent(w, r, file, fi.ModTime(), f)
	})
}

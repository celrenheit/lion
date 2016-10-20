package middleware

import (
	"fmt"
	"net/http"
	"time"
)

type MaxAge struct {
	Duration time.Duration
	Filter   func(r *http.Request) bool
}

// MaxAge is a middleware that defines the max duration headers
func (m *MaxAge) ServeNext(next http.Handler) http.Handler {
	if m.Filter == nil {
		// Filter nothing
		m.Filter = func(r *http.Request) bool { return false }
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		if !m.Filter(r) {
			return
		}
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", int(m.Duration.Seconds())))
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

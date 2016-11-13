package middleware

import (
	"net/http"
	"time"

	"github.com/celrenheit/lion"
)

type NoCache struct {
	ResponseHeaders map[string]string
	EtagHeaders     []string
}

// NoCache middleware sets headers to disable browser caching.
// Inspired by https://github.com/mytrile/nocache
func NewNoCache() lion.Middleware {
	var epoch = time.Unix(0, 0).Format(time.RFC1123)

	return NoCache{
		ResponseHeaders: map[string]string{
			"Expires":         epoch,
			"Cache-Control":   "no-cache, private, must-revalidate, max-age=0",
			"Pragma":          "no-cache",
			"X-Accel-Expires": "0",
		},
		EtagHeaders: []string{
			"ETag",
			"If-Modified-Since",
			"If-Match",
			"If-Note-Match",
			"If-Range",
			"If-Unmodified-Since",
		},
	}
}

func (n NoCache) ServeNext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delete ETag headers
		for _, v := range n.EtagHeaders {
			if r.Header.Get(v) != "" {
				r.Header.Del(v)
			}
		}

		// Add nocache headers
		for k, v := range n.ResponseHeaders {
			w.Header().Set(k, v)
		}

	})
}

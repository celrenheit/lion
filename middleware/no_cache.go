package middleware

import "net/http"

type NoCache struct {
	ResponseHeaders map[string]string
	EtagHeaders     []string
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

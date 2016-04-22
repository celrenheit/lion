// +build go1.7

package lion

import (
	"net/http"

	"context"
)

func contextFromRequest(r *http.Request) context.Context {
	return r.Context()
}

func addContextToRequest(r *http.Request, c context.Context) *http.Request {
	return r.WithContext(c)
}

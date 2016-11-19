package middleware

import (
	"context"
	"net/http"

	"github.com/nats-io/nuid"
)

type ctxRequestIDKeyType int

const CtxRequestIDKey int = 0

const (
	headerXRequestID = "X-Request-ID"
)

type RequestID struct {
	SetHeader bool
	n         *nuid.NUID
}

func NewRequestID() *RequestID {
	return &RequestID{
		SetHeader: true,
		n:         nuid.New(),
	}
}

func (rid *RequestID) ServeNext(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		requestid := rid.n.Next()
		if rid.SetHeader {
			w.Header().Set(headerXRequestID, requestid)
		}
		ctx := context.WithValue(r.Context(), CtxRequestIDKey, requestid)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

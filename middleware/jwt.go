package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/celrenheit/lion"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

var (
	DefaultJWTContextKey   = "user"
	errJWTTokenExpired     = errors.New("Token expired")
	errJWTTokenNotValidYet = errors.New("Token not valid yet")
	errJWTTokenMalformed   = errors.New("Token malformed")
	errJWTWrongAlg         = errors.New("Wrong algorithm")
)

type JWT struct {
	SigningKey    interface{}
	SigningMethod string
	ContextKey    interface{}
}

func NewJWT(secret []byte) *JWT {
	return &JWT{SigningKey: secret, SigningMethod: "HS256", ContextKey: DefaultJWTContextKey}
}

func (j *JWT) ServeNext(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c := lion.C(r)

		if h := r.Header.Get("Authorization"); h != "" {
			// Parse token from request
			token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
				if token.Method.Alg() != j.SigningMethod {
					return nil, errJWTWrongAlg
				}
				return j.SigningKey, nil
			})

			// Respond with error: 400 Bad Request
			if err != nil {
				c.WithStatus(http.StatusBadRequest)

				if vErr, ok := err.(*jwt.ValidationError); ok {

					switch vErr.Errors {
					case jwt.ValidationErrorExpired:
						err = errJWTTokenExpired
					case jwt.ValidationErrorNotValidYet:
						err = errJWTTokenNotValidYet
					case jwt.ValidationErrorMalformed:
						err = errJWTTokenMalformed
					default:
						err = vErr
					}

					c.Error(err)
					return
				}

				c.Error(err)
				return
			}

			// Token invalid respond with error: 401 Unauthorized
			if !token.Valid {
				c.Error(lion.ErrorUnauthorized)
				return
			}

			// Adding token to context key and continue to next handler
			ctx := r.Context()
			ctx = context.WithValue(ctx, j.ContextKey, token.Claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			// If there is no Authorization header then continue to the next handler
			next.ServeHTTP(w, r)
		}
	}

	return http.HandlerFunc(fn)
}

func (j *JWT) EnsureAuthenticated() lion.Middleware {
	return JWTEnsureAuthenticated(j.ContextKey)
}

func JWTEnsureAuthenticated(contextKey interface{}) lion.Middleware {
	mw := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			c := lion.C(r)
			if r.Context().Value(contextKey) == nil {
				c.Error(lion.ErrorUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	return lion.MiddlewareFunc(mw)
}

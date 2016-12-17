package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/celrenheit/htest"
	"github.com/celrenheit/lion"
	jwt "github.com/dgrijalva/jwt-go"
)

func TestJWT(t *testing.T) {
	r := newTestJWTRouter("HS256", "secret")
	test := htest.New(t, r)

	test.Get("/private").Do().
		ExpectStatus(401)

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
	test.Get("/private").
		SetHeader("Authorization", "Bearer "+token).Do().
		ExpectStatus(200)

	invalidAlg := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.EkN-DOsnsuRjRO6BxXemmJDm3HbxrbRzXglbN2S4sOkopdU4IsDxTI8jO19W_A4K8ZPJijNLis4EZsHeY559a4DFOd50_OqgHGuERTqYZyuhtF39yxJPAjUESwxk2J5k_4zM3O-vtd1Ghyo4IbqKKSy6J9mTniYJPenn5-HIirE"
	test.Get("/private").
		SetHeader("Authorization", "Bearer "+invalidAlg).Do().
		ExpectStatus(400).
		ExpectBody(errJWTWrongAlg.Error())

	// Invalid key
	r = newTestJWTRouter("HS256", "invalidsecret")
	test = htest.New(t, r)

	test.Get("/private").
		SetHeader("Authorization", "Bearer "+token).Do().
		ExpectStatus(400).
		ExpectBody(jwt.ErrSignatureInvalid.Error())

	// Malformed token
	r = newTestJWTRouter("HS256", "secret")
	test = htest.New(t, r)

	test.Get("/private").
		SetHeader("Authorization", "Bearer test").Do().
		ExpectStatus(400).
		ExpectBody(errJWTTokenMalformed.Error())

	// Expired token
	r = newTestJWTRouter("HS256", "secret")
	test = htest.New(t, r)

	expiredToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiIiLCJpYXQiOm51bGwsImV4cCI6MTI4OTk5MTE3MSwiYXVkIjoiIiwic3ViIjoiIiwibmFtZSI6IkpvaG4gRG9lIn0.QyuKmJ3mt74i0gpzP861Eek53ksTZI4mvqojDGF8tnY"
	test.Get("/private").
		SetHeader("Authorization", "Bearer "+expiredToken).
		Do().
		ExpectStatus(400).
		ExpectBody(errJWTTokenExpired.Error())

	// Not valid yet token
	r = newTestJWTRouter("HS256", "secret")
	test = htest.New(t, r)

	notvalidyetToken := generateToken("secret", 1*time.Minute, 2*time.Minute)
	test.Get("/private").
		SetHeader("Authorization", "Bearer "+notvalidyetToken).
		Do().
		ExpectStatus(400).
		ExpectBody(errJWTTokenNotValidYet.Error())

	// RS256
	r = newTestJWTRouter("RS256", `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDdlatRjRjogo3WojgGHFHYLugdUWAY9iR3fy4arWNA1KoS8kVw33cJibXr8bvwUAUparCwlvdbH6dvEOfou0/gCFQsHUfQrSDv+MuSUMAe8jzKE4qW+jK+xQU9a03GUnKHkkle+Q0pX/g6jXZ7r1/xAK5Do2kQ+X5xK9cipRgEKwIDAQAB
-----END PUBLIC KEY-----
`)
	token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.EkN-DOsnsuRjRO6BxXemmJDm3HbxrbRzXglbN2S4sOkopdU4IsDxTI8jO19W_A4K8ZPJijNLis4EZsHeY559a4DFOd50_OqgHGuERTqYZyuhtF39yxJPAjUESwxk2J5k_4zM3O-vtd1Ghyo4IbqKKSy6J9mTniYJPenn5-HIirE"
	test = htest.New(t, r)

	test.Get("/private").
		SetHeader("Authorization", "Bearer "+token).Do().
		ExpectStatus(200)
}

func newTestJWTRouter(signingMethod string, key string) *lion.Router {
	j := &JWT{
		SigningKey:    []byte(key),
		SigningMethod: signingMethod,
		ContextKey:    "jwt_key",
	}
	if signingMethod == "RS256" {
		j.SigningKey, _ = jwt.ParseRSAPublicKeyFromPEM([]byte(key))
	}

	r := lion.New()
	r.Use(j)
	r.GetFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	r.Use(j.EnsureAuthenticated())
	r.GetFunc("/private", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	return r
}

func BenchmarkJWT(b *testing.B) {
	req := httptest.NewRequest("GET", "/private", nil)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r := newTestJWTRouter("HS256", "secret")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}

func generateToken(key string, exp time.Duration, nbf time.Duration) string {
	t := jwt.New(jwt.SigningMethodHS256)
	t.Header["typ"] = "JWT"
	claims := jwt.MapClaims{}

	claims["exp"] = time.Now().Add(exp).Unix()
	claims["nbf"] = time.Now().Add(nbf).Unix()
	claims["iat"] = time.Now().Unix()
	t.Claims = claims

	str, err := t.SignedString([]byte(key))
	if err != nil {
		log.Fatal(err)
	}
	return str
}

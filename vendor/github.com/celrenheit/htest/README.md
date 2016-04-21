# HTest [![Build Status](https://img.shields.io/travis/celrenheit/htest.svg?style=flat-square)](https://travis-ci.org/celrenheit/htest) [![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/celrenheit/htest) [![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)

A lightweight high-level abstractions for testing HTTP inspired by [supertest](https://github.com/visionmedia/supertest) built around [http.ResponseRecorder](https://golang.org/pkg/net/http/httptest/#ResponseRecorder)

![Screenshot](https://raw.githubusercontent.com/celrenheit/htest/master/screenshot.png)

# Install/Update

```
$ go get -u github.com/celrenheit/htest
```

# Usage

Let's say that when we hit `/admin` we get a status code of `401 Unauthorized`, a response body of `You are not authorized` and a header of `foo=bar`

```go
mux := http.NewServeMux()
mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("foo", "bar")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, "You are not authorized")
})
```

We can write a Test to test if our program behaves correctly

```go
func TestUnauthorized(t *testing.T) {
  // We create a new instance for HTTPTester
  // It requires a testing.T and an http.Handler as argument
	h := htest.New(t, mux)

  // We make assertions to the repsonse received
	h.Get("/admin").Do().
		ExpectHeader("foo", "bar").
		ExpectStatus(http.StatusUnauthorized).
		ExpectBody("You are not authorized")
}
```

h.Get returns a [Requester](https://godoc.org/github.com/celrenheit/htest#Requester) to be able to easily build your request.
We call the Do() to execute the request and get a [ResponseAsserter](https://godoc.org/github.com/celrenheit/htest#ResponseAsserter).

There are methods for each http methods in the [HTTPTester interface](https://godoc.org/github.com/celrenheit/htest#HTTPTester)


# Building the request

We send some arbitrary data and set a header to the request. The response should have the same body and the same header value.

```go
func TestBuildingRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("foo", r.Header.Get("foo"))
		io.Copy(w, r.Body)
	})

	test := htest.New(t, mux)

	test.Get("/path").
		// Set a header to the request
		AddHeader("foo", "barbar").
		// Send a string to the request's body
		SendString("my data").

		// Executes the request
		Do().

		// The body sent should stay the same
		ExpectBody("my data").
		// The header sent should stay the same
		ExpectHeader("foo", "barbar")
}
```


# Credits

* https://github.com/go-errors/errors
* https://github.com/stretchr/testify

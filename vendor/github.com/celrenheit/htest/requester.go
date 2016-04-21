package htest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

// Requester is responsible for building an http request. When done you should call Do() method to be able to make assertions using ResponseAsserter
type Requester interface {
	// AddHeader adds a Header to the list of headers
	AddHeader(key, value string) Requester

	// SetHeader sets the Header for the key specified
	SetHeader(key, value string) Requester

	// AddCookie adds a cookie with a key and its value to the http request
	AddCookie(key, value string) Requester

	// AddForm adds request's form key and value to the request
	AddForm(key, value string) Requester

	// SetForm sets request's form key and value to the request
	SetForm(key, value string) Requester

	// FormValues allows to set a form's custom values
	FormValues(u url.Values) Requester

	// Send sends whatever data it gets as its parameter
	// The current types are supported:
	// 		- Structs, maps, slices and arrays will be marshalled to JSON
	// 		- Strings will be converted to []byte and sent to the request body
	Send(data interface{}) Requester

	// SendBytes sets a slice de bytes as the request's body
	SendBytes(data []byte) Requester

	// SendString sets a string as the request's body
	SendString(data string) Requester

	// Do executes the request and returns a ResponseAsserter allowing to perform tests on the results of this request
	Do() ResponseAsserter
}

type requester struct {
	method, path string
	body         io.Reader
	t            testing.TB
	handler      http.Handler
	request      *http.Request
}

func (r *requester) AddHeader(key, value string) Requester {
	r.request.Header.Add(key, value)
	return r
}

func (r *requester) SetHeader(key, value string) Requester {
	r.request.Header.Set(key, value)
	return r
}

func (r *requester) AddCookie(key, value string) Requester {
	r.request.AddCookie(&http.Cookie{
		Name:  key,
		Value: value,
	})
	return r
}

func (r *requester) AddForm(key, value string) Requester {
	r.request.Form.Add(key, value)
	return r
}

func (r *requester) SetForm(key, value string) Requester {
	r.request.Form.Set(key, value)
	return r
}

func (r *requester) FormValues(u url.Values) Requester {
	r.request.Form = u
	return r
}

// TODO: There are some issues with pointers and structs
func (r *requester) Send(data interface{}) Requester {
	kind := reflect.ValueOf(data).Type().Kind()
	if kind == reflect.Ptr {
		kind = reflect.ValueOf(data).Type().Elem().Kind()
	}

	switch kind {
	case reflect.Slice, reflect.Array, reflect.Struct, reflect.Map:
		b, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		r.SendBytes(b)
	case reflect.String:
		r.SendString(data.(string))
	default:
		r.t.Error("Unknow data type to send")
		r.t.FailNow()
	}
	return r
}

func (r *requester) SendBytes(data []byte) Requester {
	r.request.Body = newBody(data)
	return r
}

func (r *requester) SendString(data string) Requester {
	r.request.Body = newBodyString(data)
	return r
}

func (r *requester) Do() ResponseAsserter {
	w := httptest.NewRecorder()
	r.handler.ServeHTTP(w, r.request)
	w.Flush()
	return NewResponseAsserter(r.t, w, r.request)
}

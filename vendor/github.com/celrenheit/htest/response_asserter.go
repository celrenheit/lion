package htest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/fatih/color"
)

const lineWidth = 40

// ResponseAsserter is responsible for making assertions based on the expected and the actual value returned from httptest.ResponseRecorder
type ResponseAsserter interface {
	// ExpectHeader triggers an error if the actual value for the key is different from the expected value
	ExpectHeader(key, expected string) ResponseAsserter

	// ExpectCookie triggers an error if the actual value for the key is different from the expected value
	ExpectCookie(key, expected string) ResponseAsserter

	// ExpectCookie triggers an error if the actual value for the key is different from the expected value
	ExpectStatus(expected int) ResponseAsserter

	// ExpectBody triggers an error if actual body received is different from the expected one
	ExpectBody(expected string) ResponseAsserter

	// ExpectBody triggers an error if actual body does not contain the passed string
	ExpectBodyContains(str string) ResponseAsserter

	// ExpectBodyBytes triggers an error if actual body received is different from the expected one
	ExpectBodyBytes(b []byte) ResponseAsserter

	// ExpectJSON triggers an error if actual body received is different from the expected one.
	// Before comparing it marshals the data passed a argument using json.Marshal
	ExpectJSON(data interface{}) ResponseAsserter

	// Recorder returns the underlying ResponseRecorder instance
	Recorder() *httptest.ResponseRecorder
}

type responseAsserter struct {
	t            testing.TB
	w            *httptest.ResponseRecorder
	r            *http.Request
	printRequest sync.Once
}

// NewResponseAsserter create a new response asserter
func NewResponseAsserter(t testing.TB, w *httptest.ResponseRecorder, r *http.Request) ResponseAsserter {
	ra := &responseAsserter{
		t: t,
		w: w,
		r: r,
	}

	return ra
}

func (ra *responseAsserter) ExpectCookie(key, expected string) ResponseAsserter {
	cookies, ok := ra.w.HeaderMap["Set-Cookie"]
	if !ok {
		ra.Errorf("No cookies set")
		return ra
	}

	found := false
	for _, cookiestr := range cookies {
		splitted := strings.Split(cookiestr, "=")
		k, actual := splitted[0], splitted[1]
		if k == key {
			found = true
			if actual != expected {
				ra.ErrorKV("cookie", key, "equal", expected, actual)
			}
		}
	}

	if !found {
		ra.Errorf("Cookie %s not found", key)
	}

	return ra
}

func (ra *responseAsserter) ExpectHeader(key, expected string) ResponseAsserter {
	actual := ra.w.Header().Get(key)
	if actual != expected {
		ra.ErrorKV("header", key, "equal", expected, actual)
	}
	return ra
}

func (ra *responseAsserter) ExpectBody(expected string) ResponseAsserter {
	actual := ra.w.Body.String()
	if actual != expected {
		ra.Error("body", "equal", expected, actual)
	}
	return ra
}

func (ra *responseAsserter) ExpectBodyBytes(expected []byte) ResponseAsserter {
	actual := ra.w.Body.Bytes()
	if !bytes.Equal(actual, expected) {
		ra.Error("body", "equal", expected, actual)
	}
	return ra
}

func (ra *responseAsserter) ExpectBodyContains(expected string) ResponseAsserter {
	actual := ra.w.Body.String()
	if !strings.Contains(actual, expected) {
		ra.Error("body", "contain", expected, actual)
	}
	return ra
}

func (ra *responseAsserter) ExpectJSON(data interface{}) ResponseAsserter {
	expected, err := json.Marshal(data)
	if err != nil {
		ra.Errorf("ExpectJSON error marshalling data: %s", err.Error())
	}
	actual := ra.w.Body.Bytes()
	if !bytes.Equal(actual, expected) {
		ra.Error("JSON", "equal", string(expected), string(actual))
	}
	return ra
}

func (ra *responseAsserter) ExpectStatus(expected int) ResponseAsserter {
	actual := ra.w.Code
	if actual != expected {
		ra.Error("status code", "equal", expected, actual)
	}
	return ra
}

func (ra *responseAsserter) Error(kind, verb string, expected, actual interface{}) {
	ra.Errorf(ra.errorFormatterKV(kind, "", verb, expected, actual))
}

func (ra *responseAsserter) ErrorKV(kind, key, verb string, expected, actual interface{}) {
	ra.Errorf(ra.errorFormatterKV(kind, key, verb, expected, actual))
}

func (ra *responseAsserter) errorFormatterKV(kind, key, verb string, expected, actual interface{}) string {
	expected = wrapWithQuotesForString(expected)
	actual = wrapWithQuotesForString(actual)

	if key != "" {
		key += " "
	}

	return fmt.Sprintf("%s %sshould %s %s but got %s", magenta(strings.Title(kind)), cyan(key), verb, green(expected), red(actual))
}

func (ra *responseAsserter) Recorder() *httptest.ResponseRecorder {
	return ra.w
}

func wrapWithQuotesForString(i interface{}) interface{} {
	switch i.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, i.(string))
	default:
		return i
	}
}

var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var magenta = color.New(color.FgHiMagenta).SprintFunc()
var cyan = color.New(color.FgHiCyan).SprintFunc()

var methodColor = color.New(color.FgMagenta).SprintFunc()
var pathColor = color.New(color.Bold, color.Italic, color.FgHiBlue).SprintFunc()

func (ra *responseAsserter) Errorf(format string, args ...interface{}) {

	trace := Trace().OnlyTests().String()
	whitespaced := "\r" + getWhitespaceString() + "\r\t"

	var request string
	ra.printRequest.Do(func() {
		request = getWhitespaceString() + "\r\t" + methodColor(ra.r.Method) + " " + pathColor(ra.r.URL.Path) + "\n\r\t"
		request += red(strings.Repeat("\u2500", lineWidth))
		request += "\n\r\t"
	})
	errorStr := trace
	str := fmt.Sprintf(whitespaced+request+format+"\n\r\t"+errorStr, args...)

	ra.t.Error(str)
}

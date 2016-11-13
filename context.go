package lion

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Context key to store *ctx
var ctxKey = &struct{}{}

var (
	ErrInvalidRedirectStatusCode = errors.New("Invalid redirect status code")

	contentTypeJSON      = "application/json; charset=utf-8"
	contentTypeXML       = "application/xml; charset=utf-8"
	contentTypeTextPlain = "text/plain; charset=utf-8"
	contentTypeTextHTML  = "text/html; charset=utf-8"
)

// Check Context implements net.Context
var _ context.Context = (*ctx)(nil)
var _ Context = (*ctx)(nil)
var _ http.ResponseWriter = (*ctx)(nil)

type Context interface {
	context.Context
	http.ResponseWriter
	Param(key string) string
	ParamOk(key string) (string, bool)
	Clone() Context

	Request() *http.Request

	// Request
	Cookie(name string) (*http.Cookie, error)
	Query(name string) string
	GetHeader(key string) string

	// Response
	WithStatus(code int) Context
	WithHeader(key, value string) Context
	WithCookie(cookie *http.Cookie) Context

	// Rendering
	JSON(data interface{}) error
	XML(data interface{}) error
	String(format string, a ...interface{}) error
	File(path string) error
	Attachment(path, filename string) error
	Redirect(urlStr string) error
}

// Context implements context.Context and stores values of url parameters
type ctx struct {
	context.Context
	http.ResponseWriter

	parent context.Context
	req    *http.Request

	params []parameter

	code          int
	statusWritten bool
}

// newContext creates a new context instance
func newContext() *ctx {
	return newContextWithParent(context.Background())
}

// newContextWithParent creates a new context with a parent context specified
func newContextWithParent(c context.Context) *ctx {
	return newContextWithResReq(c, nil, nil)
}

func newContextWithResReq(c context.Context, w http.ResponseWriter, r *http.Request) *ctx {
	return &ctx{
		parent:         c,
		ResponseWriter: w,
		req:            r,
	}
}

// Value returns the value for the passed key. If it is not found in the url params it returns parent's context Value
func (c *ctx) Value(key interface{}) interface{} {
	if key == ctxKey {
		return c
	}

	if k, ok := key.(string); ok {
		if val, exist := c.ParamOk(k); exist {
			return val
		}
	}

	return c.parent.Value(key)
}

func (c *ctx) AddParam(key, val string) {
	c.params = append(c.params, parameter{key, val})
}

// Param returns the value of a param.
// If it does not exist it returns an empty string
func (c *ctx) Param(key string) string {
	val, _ := c.ParamOk(key)
	return val
}

// ParamOk returns the value of a param and a boolean that indicates if the param exists.
func (c *ctx) ParamOk(key string) (string, bool) {
	for _, p := range c.params {
		if p.key == key {
			return p.val, true
		}
	}

	return "", false
}

func (c *ctx) Clone() Context {
	nc := newContext()
	nc.parent = c.parent
	nc.params = make([]parameter, len(c.params), cap(c.params))
	copy(nc.params, c.params)

	// shallow copy of request
	nr := &c.req
	nc.req = *nr

	return nc
}

///////////// REQUEST UTILS ////////////////

func (c *ctx) Request() *http.Request {
	return c.req
}

func (c *ctx) Cookie(name string) (*http.Cookie, error) {
	return c.Request().Cookie(name)
}

func (c *ctx) Query(name string) string {
	return c.urlQueries().Get(name)
}

func (c *ctx) urlQueries() url.Values {
	return c.Request().URL.Query()
}

func (c *ctx) GetHeader(key string) string {
	return c.Request().Header.Get(key)
}

///////////// REQUEST UTILS ////////////////

///////////// RESPONSE MODIFIERS /////////////

// WithStatus sets the status code for the current request.
// If the status has already been written it will not change the current status code
func (c *ctx) WithStatus(code int) Context {
	c.code = code
	return c
}

func (c *ctx) writeHeader() {
	if !c.isStatusWritten() {
		c.WriteHeader(c.code)
		c.statusWritten = true
	}
}

func (c *ctx) isStatusWritten() bool {
	return c.statusWritten
}

// WithHeader is a convenient alias for http.ResponseWriter.Header().Set()
func (c *ctx) WithHeader(key, value string) Context {
	c.Header().Set(key, value)
	return c
}

func (c *ctx) WithCookie(cookie *http.Cookie) Context {
	http.SetCookie(c, cookie)
	return c
}

///////////// RESPONSE MODIFIERS /////////////

///////////// RESPONSE RENDERING /////////////

func (c *ctx) JSON(data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.raw(b, contentTypeJSON)
}

func (c *ctx) String(format string, a ...interface{}) error {
	str := fmt.Sprintf(format, a...)
	return c.raw([]byte(str), contentTypeTextPlain)
}

func (c *ctx) XML(data interface{}) error {
	b, err := xml.Marshal(data)
	if err != nil {
		return err
	}
	return c.raw(b, contentTypeXML)
}

func (c *ctx) File(path string) error {
	http.ServeFile(c, c.Request(), path)
	return nil
}

func (c *ctx) Attachment(path, filename string) error {
	return c.WithHeader("Content-Disposition", "attachment; filename="+filename).
		File(path)
}

func (c *ctx) setContentType(ctype string) {
	c.Header().Set("Content-Type", ctype)
}

func (c *ctx) raw(b []byte, contentType string) error {
	c.setContentType(contentType)
	c.writeHeader()
	_, err := c.Write(b)
	return err
}

///////////// RESPONSE RENDERING /////////////

func (c *ctx) Redirect(urlStr string) error {
	if c.code < 300 || c.code > 308 {
		return ErrInvalidRedirectStatusCode
	}

	http.Redirect(c, c.Request(), urlStr, c.code)
	return nil
}

func (c *ctx) Reset() {
	c.params = c.params[:0]
	c.parent = nil
	c.req = nil
	c.ResponseWriter = nil
}

func (c *ctx) Remove(key string) {
	i := c.indexOf(key)
	if i < 0 {
		panicl("Cannot remove unknown key '%s' from context", key)
	}

	c.params = append(c.params[:i], c.params[i+1:]...)
}

func (c *ctx) indexOf(key string) int {
	for i := len(c.params) - 1; i >= 0; i-- {
		if c.params[i].key == key {
			return i
		}
	}
	return -1
}

// C returns a Context based on a context.Context passed. If it does not convert to Context, it creates a new one with the context passed as argument.
func C(req *http.Request) Context {
	c := req.Context()
	if val := c.Value(ctxKey); val != nil {
		if ctx, ok := val.(*ctx); ok {
			return ctx
		}
	}
	return nil
}

// Param returns the value of a url param base on the passed context
func Param(req *http.Request, key string) string {
	return C(req).Param(key)
}

func setParamContext(req *http.Request, c *ctx) *http.Request {
	c.parent = req.Context()
	return req.WithContext(c)
}

type parameter struct {
	key string
	val string
}

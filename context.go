package lion

import (
	"context"
	"net/http"
)

// Context key to store *ctx
var ctxKey = &struct{}{}

// Check Context implements net.Context
var _ context.Context = (*ctx)(nil)
var _ Context = (*ctx)(nil)

type Context interface {
	context.Context
	Param(key string) string
	ParamOk(key string) (string, bool)
}

// Context implements context.Context and stores values of url parameters
type ctx struct {
	context.Context
	parent context.Context

	keys   []string
	values []string
}

// newContext creates a new context instance
func newContext() *ctx {
	return newContextWithParent(context.Background())
}

// newContextWithParent creates a new context with a parent context specified
func newContextWithParent(c context.Context) *ctx {
	return &ctx{
		parent: c,
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
	c.keys = append(c.keys, key)
	c.values = append(c.values, val)
}

// Param returns the value of a param.
// If it does not exist it returns an empty string
func (c *ctx) Param(key string) string {
	val, _ := c.ParamOk(key)
	return val
}

// ParamOk returns the value of a param and a boolean that indicates if the param exists.
func (c *ctx) ParamOk(key string) (string, bool) {
	for i, name := range c.keys {
		if name == key {
			return c.values[i], true
		}
	}

	if c, ok := c.parent.(*ctx); ok {
		return c.ParamOk(key)
	} else if val, ok := c.parent.Value(key).(string); ok {
		return val, ok
	}

	return "", false
}

func (c *ctx) Reset() {
	c.keys = c.keys[:0]
	c.values = c.values[:0]
	c.parent = nil
}

func (c *ctx) Remove(key string) {
	i := c.indexOf(key)
	if i < 0 {
		panicl("Cannot remove unknown key '%s' from context", key)
	}

	c.keys = append(c.keys[:i], c.keys[i+1:]...)
	c.values = append(c.values[:i], c.values[i+1:]...)
}

func (c *ctx) indexOf(key string) int {
	for i := len(c.keys) - 1; i >= 0; i-- {
		if c.keys[i] == key {
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
	return newContextWithParent(c)
}

// Param returns the value of a url param base on the passed context
func Param(req *http.Request, key string) string {
	return C(req).Param(key)
}

func setParamContext(req *http.Request, c *ctx) *http.Request {
	c.parent = req.Context()
	return req.WithContext(c)
}

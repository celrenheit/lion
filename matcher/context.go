package matcher

import "golang.org/x/net/context"

// Check Context implements net.Context

// Context implements golang.org/x/net/context.Context and stores values of url parameters
type Context interface {
	context.Context
	Param(key string) string
	ParamOk(key string) (string, bool)
	AddParam(key, value string)
	Remove(key string)
	Reset()
}

type ctx struct {
	context.Context
	parent context.Context

	params []Parameter
}

// NewContext creates a new context instance
func NewContext() Context {
	return NewContextWithParent(context.Background())
}

// NewContextWithParent creates a new context with a parent context specified
func NewContextWithParent(c context.Context) Context {
	return &ctx{
		parent: c,
	}
}

// Value returns the value for the passed key. If it is not found in the url params it returns parent's context Value
func (p *ctx) Value(key interface{}) interface{} {
	if k, ok := key.(string); ok {
		if val, exist := p.ParamOk(k); exist {
			return val
		}
	}

	return p.parent.Value(key)
}

// Param returns the value of a param.
// If it does not exist it returns an empty string
func (p *ctx) Param(key string) string {
	val, _ := p.ParamOk(key)
	return val
}

// ParamOk returns the value of a param and a boolean that indicates if the param exists.
func (p *ctx) ParamOk(key string) (string, bool) {
	for _, p := range p.params {
		if p.Key == key {
			return p.Val, true
		}
	}

	if c, ok := p.parent.(Context); ok {
		return c.ParamOk(key)
	} else if val, ok := p.parent.Value(key).(string); ok {
		return val, ok
	}

	return "", false
}

func (p *ctx) toMap() map[string]string {
	m := map[string]string{}
	for _, p := range p.params {
		m[p.Key] = p.Val
	}
	return m
}

func (p *ctx) Reset() {
	p.params = p.params[:0]
	p.parent = nil
}

func (p *ctx) Remove(key string) {
	i := p.indexOf(key)
	if i < 0 {
		panicm("Cannot remove unknown key '%s' from context", key)
	}

	p.params = append(p.params[:i], p.params[i+1:]...)
}

func (p *ctx) indexOf(key string) int {
	for i := len(p.params) - 1; i >= 0; i-- {
		if p.params[i].Key == key {
			return i
		}
	}
	return -1
}

func (p *ctx) Params() []Parameter {
	return p.params
}

func (p *ctx) AddParam(key, value string) {
	p.params = append(p.params, Parameter{Key: key, Val: value})
}

// C returns a Context based on a context.Context passed. If it does not convert to Context, it creates a new one with the context passed as argument.
func C(c context.Context) Context {
	if ctx, ok := c.(Context); ok {
		return ctx
	}
	return NewContextWithParent(c)
}

// Param returns the value of a url param base on the passed context
func Param(c context.Context, key string) string {
	return C(c).Param(key)
}

type Parameter struct {
	Key string
	Val string
}

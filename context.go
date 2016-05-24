package lion

import "golang.org/x/net/context"

// Check Context implements net.Context
var _ context.Context = (*Context)(nil)

// type ContextI interface {
// 	context.Context
// 	Param(string) string
// }

// Context implements golang.org/x/net/context.Context and stores values of url parameters
type Context struct {
	context.Context
	parent context.Context

	keys   []string
	values []string
}

// NewContext creates a new context instance
func NewContext() *Context {
	return NewContextWithParent(context.Background())
}

// NewContextWithParent creates a new context with a parent context specified
func NewContextWithParent(c context.Context) *Context {
	return &Context{
		parent: c,
	}
}

// Value returns the value for the passed key. If it is not found in the url params it returns parent's context Value
func (p *Context) Value(key interface{}) interface{} {
	if k, ok := key.(string); ok {
		if val, exist := p.ParamOk(k); exist {
			return val
		}
	}

	return p.parent.Value(key)
}

func (p *Context) addParam(key, val string) {
	p.keys = append(p.keys, key)
	p.values = append(p.values, val)
}

// Param returns the value of a param.
// If it does not exist it returns an empty string
func (p *Context) Param(key string) string {
	val, _ := p.ParamOk(key)
	return val
}

// ParamOk returns the value of a param and a boolean that indicates if the param exists.
func (p *Context) ParamOk(key string) (string, bool) {
	for i, name := range p.keys {
		if name == key {
			return p.values[i], true
		}
	}

	if c, ok := p.parent.(*Context); ok {
		return c.ParamOk(key)
	} else if val, ok := p.parent.Value(key).(string); ok {
		return val, ok
	}

	return "", false
}

func (p *Context) toMap() M {
	m := M{}
	for i := range p.keys {
		m[p.keys[i]] = p.values[i]
	}
	return m
}

func (p *Context) reset() {
	p.keys = p.keys[:0]
	p.values = p.values[:0]
	p.parent = nil
}

func (p *Context) delete(key string) {
	i := p.indexOf(key)
	if i < 0 {
		panicl("Cannot remove unknown key '%s' from context", key)
	}

	p.keys = append(p.keys[:i], p.keys[i+1:]...)
	p.values = append(p.values[:i], p.values[i+1:]...)
}

func (p *Context) indexOf(key string) int {
	for i := len(p.keys) - 1; i >= 0; i-- {
		if p.keys[i] == key {
			return i
		}
	}
	return -1
}

// C returns a Context based on a context.Context passed. If it does not convert to Context, it creates a new one with the context passed as argument.
func C(c context.Context) *Context {
	if ctx, ok := c.(*Context); ok {
		return ctx
	}
	return NewContextWithParent(c)
}

// Param returns the value of a url param base on the passed context
func Param(c context.Context, key string) string {
	return C(c).Param(key)
}

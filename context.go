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

// Param returns the value of a param
func (p *Context) Param(key string) string {
	val, _ := p.ParamOk(key)
	return val
}

func (p *Context) ParamOk(key string) (string, bool) {
	for i, name := range p.keys {
		if name == key {
			return p.values[i], true
		}
	}
	return "", false
}

func (p *Context) reset() {
	p.keys = p.keys[:0]
	p.values = p.values[:0]
	p.parent = nil
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

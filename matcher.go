package lion

import (
	"net/http"
)

// RegisterMatcher registers and matches routes to Handlers
type RegisterMatcher interface {
	Register(method, pattern string, handler Handler)
	Match(*Context, *http.Request) (*Context, Handler)
}

////////////////////////////////////////////////////////////////////////////
///												RADIX 																				 ///
////////////////////////////////////////////////////////////////////////////

var _ RegisterMatcher = (*radixMatcher)(nil)

type radixMatcher struct {
	root *node
}

func newRadixMatcher() *radixMatcher {
	r := &radixMatcher{
		root: &node{},
	}
	return r
}

func (d *radixMatcher) Register(method, pattern string, handler Handler) {
	if len(pattern) == 0 || pattern[0] != '/' {
		panic("path must begin with '/' in path '" + pattern + "'")
	}

	if d.root == nil {
		d.root = &node{}
	}

	d.root.addRoute(method, pattern, handler)
}

func (d *radixMatcher) Match(c *Context, r *http.Request) (*Context, Handler) {
	n, c := d.root.findNode(c, r.Method, cleanPath(r.URL.Path))
	if n == nil {
		return c, nil
	}
	return c, n.getHandler(r.Method)
}

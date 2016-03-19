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
	trees map[string]*node
}

func newRadixMatcher() *radixMatcher {
	r := &radixMatcher{
		trees: make(map[string]*node),
	}
	return r
}

func (d *radixMatcher) Register(method, pattern string, handler Handler) {
	if len(pattern) == 0 || pattern[0] != '/' {
		panic("path must begin with '/' in path '" + pattern + "'")
	}

	if d.trees == nil {
		d.trees = make(map[string]*node)
	}

	root := d.trees[method]
	if root == nil {
		root = &node{}
		d.trees[method] = root
	}
	root.addRoute(pattern, handler)
}

func (d *radixMatcher) Match(c *Context, r *http.Request) (*Context, Handler) {
	if root, ok := d.trees[r.Method]; ok {
		n, c := root.findNode(c, cleanPath(r.URL.Path))
		if n == nil {
			return c, nil
		}
		return c, n.handler
	}
	return c, nil
}

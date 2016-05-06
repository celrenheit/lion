package lion

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"
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
		if r.Method == OPTIONS {
			return c, d.automaticOptionsHandler(c, r.URL.Path)
		}
		return c, nil
	}
	return c, n.getHandler(r.Method)
}

func (d *radixMatcher) automaticOptionsHandler(c *Context, path string) Handler {
	allowed := make([]string, 0, len(allowedHTTPMethods))
	var fn *node // Node already found (to avoid calling too many times findNode)
	for _, method := range allowedHTTPMethods {
		if method == OPTIONS {
			continue
		}

		if fn == nil {
			n, _ := d.root.findNode(c, method, path)
			if n != nil {
				fn = n
			} else {
				continue
			}
		}

		if fn.isLeafForMethod(method) {
			allowed = append(allowed, method)
		}
	}

	if len(allowed) == 0 { // There is no method allowed
		return nil
	}

	allowed = append(allowed, OPTIONS)

	joined := strings.Join(allowed, ",")
	return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept", joined)
		w.WriteHeader(http.StatusOK)
	})
}

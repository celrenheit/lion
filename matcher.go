package lion

import "net/http"

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
	d.prevalidation(method, pattern)

	if d.root == nil {
		d.root = &node{}
	}

	d.root.addRoute(method, pattern, handler)

	d.postvalidation(method, pattern)
}

func (d *radixMatcher) Match(c *Context, r *http.Request) (*Context, Handler) {
	n, c := d.root.findNode(c, r.Method, cleanPath(r.URL.Path))
	if n == nil {
		return c, nil
	}
	return c, n.getHandler(r.Method)
}

func (d *radixMatcher) prevalidation(method, pattern string) {
	if len(pattern) == 0 || pattern[0] != '/' {
		panicl("path must begin with '/' in path '" + pattern + "'")
	}

	// Is http method allowed
	if !isInStringSlice(allowedHTTPMethods[:], method) {
		panicl("lion: invalid http method => %s\n\tShould be one of %v", method, allowedHTTPMethods)
	}
}

func (d *radixMatcher) postvalidation(method, pattern string) {
	// Find duplicate parameter names
	d.findDuplicateParamNames(d.root, method, pattern, []string{})
}

func (d *radixMatcher) findDuplicateParamNames(n *node, method, pattern string, pnames []string) {
	for _, children := range n.children {
		for _, child := range children {
			if child.nodeType > static && child.pname == "" {
				panicl(`cannot use an unnamed parameter for  %s`, pattern)
			}

			if len(child.pname) > 0 && isInStringSlice(pnames, child.pname) {
				panicl("lion: Duplicate parameter %s for %s", child.pname, pattern)
			}

			d.findDuplicateParamNames(child, method, pattern, append(pnames, child.pname))
		}
	}
}

func isInStringSlice(slice []string, expected string) bool {
	for _, val := range slice {
		if val == expected {
			return true
		}
	}
	return false
}

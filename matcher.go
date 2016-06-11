package lion

import (
	"net/http"
	"sync"

	"github.com/celrenheit/pmatch"
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
	root     *node
	matcher  pmatch.Matcher
	tagsPool sync.Pool
}

func newRadixMatcher() *radixMatcher {
	cfg := &pmatch.Config{
		ParamChar:        ':',
		WildcardChar:     '*',
		Separators:       "/.",
		GetSetterCreator: &creator{},
	}

	r := &radixMatcher{
		root:    &node{},
		matcher: pmatch.Custom(cfg),
	}
	return r
}

func (d *radixMatcher) Register(method, pattern string, handler Handler) {
	d.prevalidation(method, pattern)

	if d.root == nil {
		d.root = &node{}
	}

	d.matcher.Set(pattern, handler, pmatch.Tags{method})

	d.postvalidation(method, pattern)
}

func (d *radixMatcher) Match(c *Context, r *http.Request) (*Context, Handler) {
	p := cleanPath(r.URL.Path)

	ti := grabTagsItem()
	ti.tags = append(ti.tags, r.Method)

	h := d.matcher.GetWithContext(c, p, ti.tags)

	putTagsItem(ti)

	if handler, ok := h.(Handler); ok {
		return c, handler
	}

	return c, nil
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

func (gs *methodsHandlers) Set(value interface{}, tags pmatch.Tags) {
	if len(tags) != 1 {
		panicl("Length != 1")
	}

	method := tags[0]

	var handler Handler
	if value == nil {
		handler = nil
	} else {
		if h, ok := value.(Handler); !ok {
			panicl("Not handler")
		} else {
			handler = h
		}
	}

	gs.addHandler(method, handler)
}

func (gs *methodsHandlers) Get(tags pmatch.Tags) interface{} {
	if len(tags) != 1 {
		panicl("No method")
	}

	method := tags[0]

	return gs.getHandler(method)
}

func (gs *methodsHandlers) addHandler(method string, handler Handler) {
	switch method {
	case GET:
		gs.get = handler
	case HEAD:
		gs.head = handler
	case POST:
		gs.post = handler
	case PUT:
		gs.put = handler
	case DELETE:
		gs.delete = handler
	case TRACE:
		gs.trace = handler
	case OPTIONS:
		gs.options = handler
	case CONNECT:
		gs.connect = handler
	case PATCH:
		gs.patch = handler
	}
}

func (gs *methodsHandlers) getHandler(method string) Handler {
	switch method {
	case GET:
		return gs.get
	case HEAD:
		return gs.head
	case POST:
		return gs.post
	case PUT:
		return gs.put
	case DELETE:
		return gs.delete
	case TRACE:
		return gs.trace
	case OPTIONS:
		return gs.options
	case CONNECT:
		return gs.connect
	case PATCH:
		return gs.patch
	default:
		return nil
	}
}

type creator struct{}

func (c *creator) New() pmatch.GetSetter {
	return &methodsHandlers{}
}

//// Tags Item
type tagsItem struct {
	tags pmatch.Tags
}

func (ti *tagsItem) reset() {
	ti.tags = ti.tags[:0]
}

var tagsItemPool = sync.Pool{
	New: func() interface{} {
		return &tagsItem{}
	},
}

func grabTagsItem() *tagsItem {
	return tagsItemPool.Get().(*tagsItem)
}

func putTagsItem(ti *tagsItem) {

	ti.reset()
	tagsItemPool.Put(ti)
}

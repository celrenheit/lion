package lion

import (
	"sort"
	"strings"
)

// HTTP methods constants
const (
	GET     = "GET"
	HEAD    = "HEAD"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	TRACE   = "TRACE"
	OPTIONS = "OPTIONS"
	CONNECT = "CONNECT"
	PATCH   = "PATCH"
)

var allowedHTTPMethods = [...]string{GET, HEAD, POST, PUT, DELETE, TRACE, OPTIONS, CONNECT, PATCH}

type nodeType uint8

const (
	static   nodeType = iota // /hello
	regexp                   // TODO: /:id(regex)
	param                    // /:id
	wildcard                 // *
)

type node struct {
	nodeType    nodeType
	pattern     string
	children    typesToNodes
	label       byte
	endinglabel byte
	handlers    *methodsHandlers
	pname       string
}

type methodsHandlers struct {
	get     Handler
	head    Handler
	post    Handler
	put     Handler
	delete  Handler
	trace   Handler
	options Handler
	connect Handler
	patch   Handler
}

func (n *node) isLeaf() bool {
	for _, m := range allowedHTTPMethods {
		if n.getHandler(m) != nil {
			return true
		}
	}
	return false
}

func (n *node) isLeafForMethod(method string) bool {
	return n.getHandler(method) != nil
}

func (n *node) addHandler(method string, handler Handler) {
	if n.handlers == nil {
		n.handlers = &methodsHandlers{}
	}
	switch method {
	case GET:
		n.handlers.get = handler
	case HEAD:
		n.handlers.head = handler
	case POST:
		n.handlers.post = handler
	case PUT:
		n.handlers.put = handler
	case DELETE:
		n.handlers.delete = handler
	case TRACE:
		n.handlers.trace = handler
	case OPTIONS:
		n.handlers.options = handler
	case CONNECT:
		n.handlers.connect = handler
	case PATCH:
		n.handlers.patch = handler
	}
}

func (n *node) getHandler(method string) Handler {
	if n.handlers == nil {
		return nil
	}
	switch method {
	case GET:
		return n.handlers.get
	case HEAD:
		return n.handlers.head
	case POST:
		return n.handlers.post
	case PUT:
		return n.handlers.put
	case DELETE:
		return n.handlers.delete
	case TRACE:
		return n.handlers.trace
	case OPTIONS:
		return n.handlers.options
	case CONNECT:
		return n.handlers.connect
	case PATCH:
		return n.handlers.patch
	default:
		return nil
	}
}

func (n *node) addRoute(method, pattern string, handler Handler) {
	search := pattern

	if len(search) == 0 {
		n.addHandler(method, handler)
		return
	}
	child := n.getEdge(search[0])
	if child == nil {
		child = &node{
			label:   search[0],
			pattern: search,
		}
		child.addHandler(method, handler)
		n.addChild(method, child)
		return
	}

	if child.nodeType > static {
		pos := stringsIndex(search, '/')
		if pos < 0 {
			pos = len(search)
		}

		///// Check conflicting param names
		var (
			xpattern string
			pname    string
		)

		xpattern = search[:pos]

		// Find parameter name
		if child.nodeType == wildcard {
			pname = "*"
			if len(xpattern) > 1 {
				pname = xpattern[1:]
			}
		} else {
			pname = xpattern[1:]
		}

		if pname != child.pname {
			panicl("Conflicting parameter name '%s' with '%s' for pattern: '%s'", child.pname, pname, n.pattern+pattern)
		}
		///// End check

		search = search[pos:]

		child.addRoute(method, search, handler)
		return
	}

	commonPrefix := child.longestPrefix(search)
	if commonPrefix == len(child.pattern) {

		search = search[commonPrefix:]

		child.addRoute(method, search, handler)
		return
	}

	subchild := &node{
		nodeType: static,
		pattern:  search[:commonPrefix],
	}

	n.replaceChild(search[0], subchild)
	c2 := child
	c2.label = child.pattern[commonPrefix]
	subchild.addChild(method, c2)
	child.pattern = child.pattern[commonPrefix:]

	search = search[commonPrefix:]
	if len(search) == 0 {
		subchild.addHandler(method, handler)
		return
	}
	tmp := &node{
		label:    search[0],
		nodeType: static,
		pattern:  search,
	}
	tmp.addHandler(method, handler)
	subchild.addChild(method, tmp)
	return
}

func (n *node) getEdge(label byte) *node {
	for _, nds := range n.children {
		for _, n := range nds {
			if n.label == label {
				return n
			}
		}
	}

	return nil
}

func (n *node) replaceChild(label byte, child *node) {
	for i := 0; i < len(n.children[child.nodeType]); i++ {
		if n.children[child.nodeType][i].label == label {
			n.children[child.nodeType][i] = child
			n.children[child.nodeType][i].label = label
			return
		}
	}

	panic("cannot replace child")
}

func (n *node) findNode(c *Context, method, path string) (*node, *Context) {
	root := n
	search := path

	// Stores the previous node, search path and the previous i
	prev := n
	prevsearch := ""
	previ := 0
	prevparam := ""

LOOP:
	for {
		if len(search) == 0 && root.children.isEmpty() {
			break
		}

		l := len(root.children)
		for i := 0; i < l; i++ {
			t := nodeType(i)

			if len(root.children[i]) == 0 {
				// If the searched path does not start with the current pattern and there are no children greather than the current nodeType.
				// Then go back to the previous(parent) node and search for childs of the next nodeType.
				if !strings.HasPrefix(search, root.pattern) && root.children.isEmptyStartingWith(t+1) && prev != root {
					root = prev
					search = prevsearch
					i = previ

					// Delete previous param
					if prevparam != "" {
						c.delete(prevparam)
					}
				}
				continue
			}

			var label byte
			if len(search) > 0 {
				label = search[0]
			}

			xn := root.findEdge(t, label)
			if xn == nil {
				continue
			}

			xsearch := search
			if xn.nodeType > static {
				p := -1
				if xn.nodeType < wildcard {
					// To match or not match . in path
					chars := "/"
					if xn.endinglabel == '.' {
						chars += "."
					}
					p = stringsIndexAny(xsearch, chars)
				}

				if p < 0 {
					p = len(xsearch)
				}

				if xn.nodeType == wildcard {
					c.addParam(xn.pname, xsearch)
				} else {
					c.addParam(xn.pname, xsearch[:p])
				}

				prevparam = xn.pname // Stores the previous param name

				xsearch = xsearch[p:]
			} else if strings.HasPrefix(xsearch, xn.pattern) {
				xsearch = xsearch[len(xn.pattern):]
			} else {
				continue
			}

			if len(xsearch) == 0 && xn.isLeafForMethod(method) {
				return xn, c
			}

			prev = root
			root = xn

			prevsearch = search
			search = xsearch

			previ = i

			continue LOOP // Search for next node (xn)
		}

		break
	}

	return nil, c
}

func (n *node) findEdge(ndtype nodeType, label byte) *node {
	nds := n.children[ndtype]
	l := len(nds)
	idx := 0

LOOP:
	for ; idx < l; idx++ {
		switch ndtype {
		case static:
			if nds[idx].label >= label {
				break LOOP
			}
		default:
			break LOOP
		}
	}

	if idx >= l {
		return nil
	}
	node := nds[idx]
	if node.nodeType == static && node.label == label {
		return node
	} else if node.nodeType > static {
		return node
	}
	return nil
}

func (n *node) isEdge() bool {
	return n.label != 0
}

func (n *node) longestPrefix(pattern string) int {
	return longestPrefix(n.pattern, pattern)
}

func (n *node) addChild(method string, child *node) {
	search := child.pattern
	pos := stringsIndexAny(search, ":*")

	ndtype := static
	if pos >= 0 {
		switch search[pos] {
		case ':':
			ndtype = param
		case '*':
			ndtype = wildcard
		}
	}

	switch {
	case pos == 0: // Pattern starts with wildcard
		l := len(search)
		handler := child.getHandler(method)
		child.nodeType = ndtype
		var (
			endingpos int
			pname     string
		)
		if ndtype == wildcard {
			endingpos = -1
		} else {
			endingpos = stringsIndexAny(search, "./")
		}
		if endingpos < 0 {
			endingpos = l
		} else {
			child.endinglabel = search[endingpos]
		}
		child.pattern = search[:endingpos]

		// Find parameter name
		if ndtype == wildcard {
			pname = "*"
			if len(child.pattern) > 1 {
				pname = child.pattern[1:]
			}
		} else {
			pname = child.pattern[1:]
		}

		child.pname = pname

		if endingpos != l {
			child.addHandler(method, nil)
			search = search[endingpos:]
			subchild := &node{
				label:    search[0],
				pattern:  search,
				nodeType: static,
			}
			subchild.addHandler(method, handler)
			child.addChild(method, subchild)
		}

	case pos > 0: // Pattern has a wildcard parameter
		handler := child.getHandler(method)

		child.nodeType = static
		child.pattern = search[:pos]
		child.addHandler(method, nil)

		search = search[pos:]
		subchild := &node{
			label:    search[0],
			nodeType: ndtype,
			pattern:  search,
		}
		subchild.addHandler(method, handler)
		child.addChild(method, subchild)
	default: // all static
		child.nodeType = ndtype
	}

	n.children[child.nodeType] = append(n.children[child.nodeType], child)
	n.children[child.nodeType].Sort()
}

type nodes []*node

func (ns nodes) Len() int           { return len(ns) }
func (ns nodes) Less(i, j int) bool { return ns[i].label < ns[j].label }
func (ns nodes) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns nodes) Sort()              { sort.Sort(ns) }

type typesToNodes [wildcard + 1]nodes

func (tps typesToNodes) isEmpty() bool {
	return tps.isEmptyStartingWith(static)
}

func (tps typesToNodes) isEmptyStartingWith(t nodeType) bool {
	for i := t; i < wildcard+1; i++ {
		if len(tps[i]) > 0 {
			return false
		}
	}
	return true
}

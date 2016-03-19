package lion

import (
	"sort"
	"strings"
)

type nodeType uint8

const (
	static   nodeType = iota // /hello
	regexp                   // TODO: /:id(regex)
	param                    // /:id
	wildcard                 // *
)

type tree struct {
	subtrees map[string]*node
}

func newTree() *tree {
	return &tree{
		subtrees: make(map[string]*node),
	}
}

func (t *tree) addRoute(method, pattern string, handler Handler) {
	root := t.subtrees[method]
	if root == nil {
		root = &node{}
		t.subtrees[method] = root
	}
	root.addRoute(pattern, handler)
}

type node struct {
	nodeType    nodeType
	pattern     string
	handler     Handler
	children    typesToNodes
	label       byte
	endinglabel byte
}

func (n *node) isLeaf() bool {
	return n.handler != nil
}

func (n *node) addRoute(pattern string, handler Handler) {
	search := pattern

	if len(search) == 0 {
		n.handler = handler
		return
	}
	child := n.getEdge(search[0])
	if child == nil {
		child = &node{
			label:   search[0],
			pattern: search,
			handler: handler,
		}
		n.addChild(child)
		return
	}

	if child.nodeType > static {
		pos := stringsIndex(search, '/')
		if pos < 0 {
			pos = len(search)
		}

		search = search[pos:]

		child.addRoute(search, handler)
		return
	}

	commonPrefix := child.longestPrefix(search)
	if commonPrefix == len(child.pattern) {

		search = search[commonPrefix:]

		child.addRoute(search, handler)
		return
	}

	subchild := &node{
		nodeType: static,
		pattern:  search[:commonPrefix],
	}

	n.replaceChild(search[0], subchild)
	c2 := child
	c2.label = child.pattern[commonPrefix]
	subchild.addChild(c2)
	child.pattern = child.pattern[commonPrefix:]

	search = search[commonPrefix:]
	if len(search) == 0 {
		subchild.handler = handler
		return
	}

	subchild.addChild(&node{
		label:    search[0],
		nodeType: static,
		pattern:  search,
		handler:  handler,
	})
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

func (n *node) findNode(c *Context, path string) (*node, *Context) {
	root := n
	search := path

LOOP:
	for {

		if len(search) == 0 {
			break
		}

		l := len(root.children)
		for i := 0; i < l; i++ {
			t := nodeType(i)

			if len(root.children[i]) == 0 {
				continue
			}

			var label byte
			if len(search) > 0 {
				label = search[0]
			}

			xn := root.findEdge(nodeType(t), label)
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
					c.addParam("*", xsearch)
				} else {
					c.addParam(xn.pattern[1:], xsearch[:p])
				}

				xsearch = xsearch[p:]
			} else if strings.HasPrefix(xsearch, xn.pattern) {
				xsearch = xsearch[len(xn.pattern):]
			} else {
				continue
			}

			if len(xsearch) == 0 && xn.isLeaf() {
				return xn, c
			}

			root = xn
			search = xsearch
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

func (n *node) addChild(child *node) {
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
		handler := child.handler
		child.nodeType = ndtype
		if ndtype == wildcard {
			pos = -1
		} else {
			pos = stringsIndexAny(search, "./")
		}
		if pos < 0 {
			pos = l
		} else {
			child.endinglabel = search[pos]
		}

		child.pattern = search[:pos]

		if pos != l {
			child.handler = nil

			search = search[pos:]
			subchild := &node{
				label:    search[0],
				pattern:  search,
				nodeType: static,
				handler:  handler,
			}
			child.addChild(subchild)
		}

	case pos > 0: // Pattern has a wildcard parameter
		handler := child.handler

		child.nodeType = static
		child.pattern = search[:pos]
		child.handler = nil

		search = search[pos:]
		subchild := &node{
			label:    search[0],
			nodeType: ndtype,
			pattern:  search,
			handler:  handler,
		}
		child.addChild(subchild)
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

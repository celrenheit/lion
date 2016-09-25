package matcher

import "sort"

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
	pname       string
	values      interface{}
	tags        Tags
	GetSetter   GetSetter
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

package matcher

import (
	"regexp"
	"sort"
	"strings"
)

type nodeType uint8

const (
	static   nodeType = iota // /hello
	param                    // /:id or /:id(regex)
	wildcard                 // *
)

type node struct {
	nodeType    nodeType
	pname       string
	re          *regexp.Regexp
	pattern     string
	label       byte
	endinglabel byte
	GetSetter   GetSetter
	priority    int

	parent *node

	staticChildren nodes
	paramChild     *node
	anyChild       *node
}

func (n *node) longestPrefix(pattern string) int {
	return longestPrefix(n.pattern, pattern)
}

func (n *node) children() nodes {
	children := make([]*node, 0, len(n.staticChildren)+2)
	for _, staticChild := range n.staticChildren {
		children = append(children, staticChild)
	}
	if n.paramChild != nil {
		children = append(children, n.paramChild)
	}
	if n.anyChild != nil {
		children = append(children, n.anyChild)
	}
	return children
}

type nodes []*node

func (ns nodes) Len() int           { return len(ns) }
func (ns nodes) Less(i, j int) bool { return ns[i].priority > ns[j].priority }
func (ns nodes) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns nodes) Sort()              { sort.Sort(ns) }
func (ns nodes) Contains(pattern, sep string) bool {
	for _, n := range ns {
		if strings.TrimSuffix(n.pattern, sep) == pattern {
			return true
		}
	}
	return false
}

func (n *node) path() string {
	if n.parent == nil {
		return n.pattern
	}

	return n.parent.path() + n.pattern
}

func (n *node) root() *node {
	if n.parent == nil {
		return n
	}

	return n.parent.root()
}

func (n *node) setStaticChild(label byte, child *node) {
	if n.staticChildren == nil {
		n.staticChildren = nodes{}
	}

	if _, ok := n.getStaticChild(label); ok {
		n.removeLabel(label)
	}

	n.staticChildren = append(n.staticChildren, child)
	n.calculatePriority()
	n.staticChildren.Sort()
}

func (n *node) removeLabel(label byte) {
	for i, c := range n.staticChildren {
		if c.label == label {
			n.staticChildren = append(n.staticChildren[:i], n.staticChildren[i+1:]...)
			return
		}
	}
	panic("Should not be accessible. If the issue persist, please report an issue.")
}

func (n *node) getStaticChild(label byte) (child *node, ok bool) {
	for _, c := range n.staticChildren {
		if c.label == label {
			return c, true
		}
	}

	return nil, false
}

func (n *node) calculatePriority() int {
	n.priority = 1
	for _, sc := range n.staticChildren {
		n.priority += sc.calculatePriority()
	}

	if n.paramChild != nil {
		n.priority += n.paramChild.calculatePriority()
	}

	if n.anyChild != nil {
		n.priority += n.anyChild.calculatePriority()
	}

	return n.priority
}

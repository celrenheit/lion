package matcher

import (
	"fmt"
	"strings"
)

type tree struct {
	root *node
	cfg  *Config
}

func (t *tree) ParamChar() byte {
	return t.cfg.ParamChar
}

func (t *tree) WildcardChar() byte {
	return t.cfg.WildcardChar
}

func (t *tree) MainSeparators() string {
	if len(t.cfg.Separators) > 0 {
		return string(t.cfg.Separators[0])
	}

	return ""
}

func (t *tree) OptionalSeparators() string {
	if len(t.cfg.Separators) > 1 {
		return t.cfg.Separators[1:]
	}

	return ""
}

func (t *tree) Separators() string {
	return t.cfg.Separators
}

func (t *tree) AllChars() string {
	return string([]byte{t.ParamChar(), t.WildcardChar()})
}

func (t *tree) setValue(n *node, value interface{}, tags Tags) {
	if t.cfg.New != nil {
		if n.GetSetter == nil {
			n.GetSetter = t.cfg.New()
		}
	}

	if n.GetSetter != nil {
		n.GetSetter.Set(value, tags)
	}
}

func newTree(cfg *Config) *tree {
	return &tree{
		root: &node{},
		cfg:  cfg,
	}
}

func (t *tree) getValue(n *node, tags Tags) interface{} {
	if n.GetSetter == nil {
		return nil
	}

	return n.GetSetter.Get(tags)
}

func (t *tree) isLeaf(n *node, tags Tags) bool {
	return t.getValue(n, tags) != nil
}

func (tree *tree) findNode(c Context, path string, tags Tags) (out *node) {
	n := tree.root
	search := path

	// Store the previous elements
	var (
		prev       *node    = n
		prevstep   nodeType = static
		prevsearch string
		prevparam  string
	)

	for {

		if search == "" && tree.isLeaf(n, tags) {
			out = n
			break
		}

		var label byte
		if search != "" {
			label = search[0]
		}

		// We check if there is a present route starting with label byte
		if nn, ok := n.getStaticChild(label); ok && stringsHasPrefix(search, nn.pattern) {
			lnn := len(nn.pattern)

			// Case where the current path starts with and is longer than the found node's (nn) static path
			// Check the tests, for example if we define:
			// 		/hello/contact/named
			// 		/hello/contact/:param
			// and the user tries to fetch:
			// 		/hello/contact/nameddd
			// it should go the second registered pattern (the one that has :param)
			if n.paramChild != nil &&
				len(search) > lnn &&
				!isByteInString(nn.endinglabel, tree.Separators()) {

				end := strings.IndexAny(search[lnn:], tree.MainSeparators())
				if end < 0 {
					end = len(search[lnn:])
				}
				end += lnn

				if search[lnn:end] != "" {
					goto PARAM
				}
			}

			n = nn
			search = search[lnn:]
			continue
		}

	PARAM:
		// If there is a param child then we go for it.
		if n.paramChild != nil {
			prev = n
			prevstep = param
			n = n.paramChild
			p := -1

			chars := tree.MainSeparators()
			if isByteInString(n.endinglabel, tree.OptionalSeparators()) {
				chars += tree.OptionalSeparators()
			}
			p = stringsIndexAny(search, chars)
			if p < 0 {
				p = len(search)
			}

			pval := tree.cfg.ParamTransformer.Transform(search[:p])
			c.AddParam(n.pname, pval)
			prevparam = n.pname
			prevsearch = search
			search = search[p:]
			continue
		}

	WILDCARD:
		// If there is a wildcard child then we go for it.
		if n.anyChild != nil {
			prev = n
			prevstep = wildcard
			n = n.anyChild

			pval := tree.cfg.ParamTransformer.Transform(search)
			c.AddParam(n.pname, pval)

			prevparam = n.pname
			prevsearch = search
			search = search[len(search):]
			continue
		}

		// Finally if we were previously in a param node and there is no matched routes.
		// We go back to the parent node and the previous search path.
		// We then jump to the parent's wildcard node.
		// If there was a previously registered param in the previous param node, we remove it.
		if n != prev && prevstep == param && prev.anyChild != nil {
			n = prev
			search = prevsearch
			if prevparam != "" {
				c.Remove(prevparam)
			}
			goto WILDCARD
		}
		break
	}

	return out
}

func (tree *tree) addRoute(n *node, pattern string, values interface{}, tags Tags) {
	splitted := tree.split(pattern)

	var cn *node
	for _, cn = range splitted {
	CONTINUE:
		switch {
		case cn.nodeType == param:
			if n.paramChild == nil {
				n.paramChild = cn
			} else {
				// Check conflicting parameter name
				if n.paramChild.pname != cn.pname {
					panicm("Conflicting parameter name '%s' with '%s' for pattern: '%s'",
						n.paramChild.pname, cn.pname, n.paramChild.path())
				}
			}

			cn.parent = n
			n = n.paramChild

			lcp := n.longestPrefix(pattern)
			pattern = pattern[lcp:]
		case cn.nodeType == wildcard:
			if n.anyChild == nil {
				n.anyChild = cn
			} else {
				// Check conflicting wildcard parameter name
				if n.anyChild.pname != cn.pname {
					panicm("Conflicting parameter name '%s' with '%s' for pattern: '%s'",
						n.anyChild.pname, cn.pname, n.anyChild.path())
				}
			}

			cn.parent = n
			n = n.anyChild

			lcp := n.longestPrefix(pattern)
			pattern = pattern[lcp:]
		default:
			fn, ok := n.getStaticChild(cn.label)
			if !ok {
				// Label does not exist in node's (n) static children
				// We then set the current node (cn) to n's static children.
				n.setStaticChild(cn.label, cn)

				cn.parent = n
				lcp := cn.longestPrefix(pattern)
				n = cn

				pattern = pattern[lcp:]

				continue
			}

			// Label already exist
			lcp := fn.longestPrefix(pattern)
			if lcp == len(fn.pattern) {
				// If the longest common prefix (lcp) between the found node (fn) and the current pattern
				// is equal to the found node's pattern.
				// Then we can use the found node as the root node (n) and continue with the next splitted node. (with one exception, see below)
				pattern = pattern[lcp:]

				fn.parent = n
				n = fn

				// If the lcp is not equal to current splitted node's length then we stay with the current splitted node (cn)
				// and adapt it's pattern and label
				if lcp != len(cn.pattern) {
					cn.pattern = cn.pattern[lcp:]
					cn.label = cn.pattern[0]
					goto CONTINUE
				}
				continue
			} else if lcp == len(cn.pattern) {
				// If the longest common prefix is equal to the current splitted node
				// we split the existing found node until the common prefix
				// and add the found node to this newly created node with it's pattern stripped.
				splitpattern := fn.pattern[:lcp]

				nfn := &node{
					parent:      n,
					pattern:     splitpattern,
					label:       splitpattern[0],
					nodeType:    static,
					endinglabel: splitpattern[len(splitpattern)-1],
				}

				if _, ok := n.getStaticChild(fn.label); ok {
					n.removeLabel(nfn.label)

					fn.pattern = fn.pattern[lcp:]
					fn.label = fn.pattern[0]

					nfn.setStaticChild(fn.label, fn)
				}

				n.setStaticChild(nfn.label, nfn)

				n = nfn
				pattern = pattern[lcp:]
				continue
			}

			// Split
			splitpattern := fn.pattern[:lcp]

			//	We create a new static node that contains the longest common prefix
			nfn := &node{
				parent:   n,
				pattern:  splitpattern,
				label:    splitpattern[0],
				nodeType: static,
			}

			n.removeLabel(nfn.label)

			// 	Then we add both the found node and splitted node with the common prefix stripped out
			if fn.pattern[lcp:] != "" {
				fn.pattern = fn.pattern[lcp:]
				fn.label = fn.pattern[0]
				nfn.setStaticChild(fn.label, fn)
				fn.parent = nfn
			}

			if cn.pattern[lcp:] != "" {
				cn.pattern = cn.pattern[lcp:]
				cn.label = cn.pattern[0]
				nfn.setStaticChild(cn.label, cn)
				cn.parent = nfn
			}

			n.setStaticChild(nfn.label, nfn)

			n = nfn
			pattern = pattern[lcp:]

			goto CONTINUE
		}
	}

	tree.setValue(n, values, tags)
}

// split splits a pattern into multiple nodes types
func (tree *tree) split(pattern string) (out []*node) {
	base := pattern
	for {
		if pattern == "" {
			break
		}
		c := pattern[0]

		var endinglabel byte

		end := strings.IndexAny(pattern, tree.Separators())
		if end < 0 {
			end = len(pattern)
			endinglabel = pattern[end-1]
		} else {
			endinglabel = pattern[end]
		}

		var child *node
		switch c {
		case tree.ParamChar():
			var l byte
			idx := strings.Index(base, pattern[:end])
			if idx > 0 {
				l = base[idx-1]
			}

			endinglabel = 0
			if len(pattern) > end {
				endinglabel = pattern[end]
			}
			child = &node{
				pattern:     pattern[:end],
				nodeType:    param,
				pname:       pattern[1:end],
				endinglabel: endinglabel,
				label:       l,
			}
			out = append(out, child)
		case tree.WildcardChar():
			pname := pattern[1:]
			if pname == "" {
				pname = "*"
			}
			child = &node{
				pattern:  pattern,
				nodeType: wildcard,
				pname:    pname,
			}

			out = append(out, child)
			pattern = ""
			continue
		default:
			charIdx := strings.IndexAny(pattern, tree.AllChars())
			if charIdx < 0 {
				charIdx = len(pattern)
			}

			end = charIdx

			cp := pattern[:end]
			endinglabel = cp[len(cp)-1]
			pattern = pattern[end:]

			child = &node{
				pattern:     cp,
				nodeType:    static,
				label:       c,
				endinglabel: endinglabel,
			}

			out = append(out, child)
			continue
		}
		pattern = pattern[end:]
	}

	return
}

func (tree *tree) printTree(n *node, decalage int) (out string) {
	dec := strings.Repeat("\t", decalage)
	out += fmt.Sprintf("%s-> %s %v ('%s' -> '%s') [%p] %d\n", dec, n.pattern, n.GetSetter != nil, string(n.label), string(n.endinglabel), n.GetSetter, n.priority)

	if len(n.staticChildren) > 0 {
		out += dec + "\tStatic Nodes\n"
	}
	for _, sc := range n.staticChildren {
		out += tree.printTree(sc, decalage+1)
	}
	if n.paramChild != nil {
		out += dec + "\tParam Node\n"
		out += tree.printTree(n.paramChild, decalage+1)
	}
	if n.anyChild != nil {
		out += dec + "\tAny Node\n"
		out += tree.printTree(n.anyChild, decalage+1)
	}
	return
}

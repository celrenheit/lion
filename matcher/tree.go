package matcher

import "strings"

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
	n.values = value
	n.tags = tags

	if t.cfg.GetSetterCreator != nil {
		if n.GetSetter == nil {
			n.GetSetter = t.cfg.GetSetterCreator.New()
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
	if n.values == nil {
		return nil
	}

	if n.GetSetter != nil {
		return n.GetSetter.Get(tags)
	} else {
		return n.values
	}
}

func (t *tree) isLeaf(n *node, tags Tags) bool {
	return t.getValue(n, tags) != nil
}

func (tree *tree) findNode(c Context, path string, tags Tags) (out *node) {
	root := tree.root
	search := path

	// Stores the previous node, search path and the previous i
	prev := tree.root
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
						c.Remove(prevparam)
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
					chars := tree.MainSeparators()

					if isByteInString(xn.endinglabel, tree.OptionalSeparators()) {
						chars += tree.OptionalSeparators()
					}
					p = strings.IndexAny(xsearch, chars)
				}

				if p < 0 {
					p = len(xsearch)
				}

				if xn.nodeType == wildcard {
					c.AddParam(xn.pname, tree.cfg.ParamTransformer.Transform(xsearch))
				} else {
					c.AddParam(xn.pname, tree.cfg.ParamTransformer.Transform(xsearch[:p]))
				}

				prevparam = xn.pname // Stores the previous param name

				xsearch = xsearch[p:]
			} else if strings.HasPrefix(xsearch, xn.pattern) {
				xsearch = xsearch[len(xn.pattern):]
			} else {
				continue
			}

			if len(xsearch) == 0 && tree.isLeaf(xn, tags) {
				return xn
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

	return nil
}

func (tree *tree) addRoute(n *node, pattern string, values interface{}, tags Tags) {
	search := pattern

	if len(search) == 0 {
		tree.setValue(n, values, tags)
		return
	}
	child := n.getEdge(search[0])
	if child == nil {
		child = &node{
			label:   search[0],
			pattern: search,
		}
		tree.setValue(child, values, tags)
		tree.addChild(n, child, values, tags)
		return
	}

	if child.nodeType > static {
		pos := strings.Index(search, tree.MainSeparators())
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
			panicm("Conflicting parameter name '%s' with '%s' for pattern: '%s'", child.pname, pname, n.pattern+pattern)
		}
		///// End check

		search = search[pos:]

		tree.addRoute(child, search, values, tags)
		return
	}

	commonPrefix := child.longestPrefix(search)
	if commonPrefix == len(child.pattern) {

		search = search[commonPrefix:]

		tree.addRoute(child, search, values, tags)
		return
	}

	subchild := &node{
		nodeType: static,
		pattern:  search[:commonPrefix],
	}

	n.replaceChild(search[0], subchild)
	c2 := child
	c2.label = child.pattern[commonPrefix]
	tree.addChild(subchild, c2, values, tags)
	child.pattern = child.pattern[commonPrefix:]

	search = search[commonPrefix:]
	if len(search) == 0 {
		tree.setValue(subchild, values, tags)
		return
	}
	tmp := &node{
		label:    search[0],
		nodeType: static,
		pattern:  search,
	}
	tree.setValue(tmp, values, tags)
	tree.addChild(subchild, tmp, values, tags)
	return
}

func (tree *tree) addChild(n *node, child *node, values interface{}, tags Tags) {
	search := child.pattern
	pos := strings.IndexAny(search, tree.AllChars())

	ndtype := static
	if pos >= 0 {
		switch search[pos] {
		case tree.ParamChar():
			ndtype = param
		case tree.WildcardChar():
			ndtype = wildcard
		}
	}

	switch {
	case pos == 0: // Pattern starts with wildcard
		l := len(search)
		handler := tree.getValue(child, tags)
		child.nodeType = ndtype
		var (
			endingpos int
			pname     string
		)
		if ndtype == wildcard {
			endingpos = -1
		} else {
			endingpos = strings.IndexAny(search, tree.Separators())
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
			tree.setValue(child, nil, tags) // TODO: create clear function
			search = search[endingpos:]
			subchild := &node{
				label:    search[0],
				pattern:  search,
				nodeType: static,
			}
			tree.setValue(subchild, handler, tags)
			tree.addChild(child, subchild, values, tags)
		}

	case pos > 0: // Pattern has a wildcard parameter
		handler := tree.getValue(child, tags)

		child.nodeType = static
		child.pattern = search[:pos]
		tree.setValue(child, nil, tags) // TODO: create clear function

		search = search[pos:]
		subchild := &node{
			label:    search[0],
			nodeType: ndtype,
			pattern:  search,
		}
		tree.setValue(subchild, handler, tags)
		tree.addChild(child, subchild, values, tags)
	default: // all static
		child.nodeType = ndtype
	}

	n.children[child.nodeType] = append(n.children[child.nodeType], child)
	n.children[child.nodeType].Sort()
}

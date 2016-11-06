package matcher

import "fmt"

type Matcher interface {
	Set(pattern string, values interface{}, tags Tags) GetSetter
	Get(pattern string, tags Tags) (Context, interface{})
	GetWithContext(c Context, pattern string, tags Tags) interface{}
	Eval(pattern string, params map[string]string) (string, error)
}

type GetSetter interface {
	Set(value interface{}, tags Tags)
	Get(tags Tags) interface{}
}

type ParamTransformer interface {
	Transform(input string) (output string)
}

type Config struct {
	ParamChar        byte
	WildcardChar     byte
	Separators       string
	ParamTransformer ParamTransformer
	New              func() GetSetter
}

type matcher struct {
	tree *tree
}

func New() Matcher {
	return Custom(&Config{
		ParamChar:        ':',
		WildcardChar:     '*',
		Separators:       "/.",
		ParamTransformer: noopParamTransformer{},
	})
}

func Custom(cfg *Config) Matcher {
	if cfg.ParamTransformer == nil {
		cfg.ParamTransformer = noopParamTransformer{}
	}

	return &matcher{
		tree: newTree(cfg),
	}
}

func (m *matcher) Set(pattern string, values interface{}, tags Tags) GetSetter {
	value := m.tree.addRoute(m.tree.root, pattern, values, tags)
	m.postvalidation(pattern)
	return value
}

func (m *matcher) Get(pattern string, tags Tags) (Context, interface{}) {
	c := NewContext()
	v := m.GetWithContext(c, pattern, tags)
	return c, v
}

func (m *matcher) GetWithContext(c Context, pattern string, tags Tags) interface{} {
	n := m.tree.findNode(c, pattern, tags)
	if n == nil {
		return nil
	}

	if !m.tree.isLeaf(n, tags) {
		return nil
	}

	return m.tree.getValue(n, tags)
}

func (m *matcher) postvalidation(pattern string) {
	// Find duplicate parameter names
	m.findDuplicateParamNames(m.tree.root, pattern, []string{})
}

func (m *matcher) findDuplicateParamNames(n *node, pattern string, pnames []string) {
	for _, sc := range n.staticChildren {
		m.findDuplicateParamNames(sc, pattern, pnames)
	}

	if n.paramChild != nil {
		nn := n.paramChild
		m.validateParamNode(nn, pattern, pnames)
		m.findDuplicateParamNames(nn, pattern, append(pnames, nn.pname))
	}

	if n.anyChild != nil {
		nn := n.anyChild
		m.validateParamNode(nn, pattern, pnames)
		m.findDuplicateParamNames(nn, pattern, append(pnames, nn.pname))
	}
}

func (m *matcher) validateParamNode(nn *node, pattern string, pnames []string) {
	if nn.pname == "" {
		panicm(`cannot use an unnamed parameter for  %s`, pattern)
	}
	if isInStringSlice(pnames, nn.pname) {
		panicm("duplicate parameter %s for %s", nn.pname, pattern)
	}
}

func (m *matcher) Eval(pattern string, params map[string]string) (string, error) {
	// TODO: Avoid .split()
	parents := m.tree.split(pattern)

	var path string
	for _, fn := range parents {
		switch fn.nodeType {
		case static:
			path += fn.pattern
		case param:
			p, ok := params[fn.pname]
			if !ok {
				return "", fmt.Errorf("Param '%s' not set", fn.pname)
			}

			if fn.re == nil {
				path += p
			} else {
				if foundStr := fn.re.FindString(p); len(foundStr) != len(p) {
					return "", fmt.Errorf("Param '%s' does not match entirely the regex pattern: '%s'", p, fn.re.String())
				}
				path += p
			}
		case wildcard:
			p, ok := params[fn.pname]
			if !ok {
				return "", fmt.Errorf("Wildcard Param '%s' not set", fn.pname)
			}
			path += p
		}
	}

	return path, nil
}

type Tags []string

type noopParamTransformer struct{}

func (_ noopParamTransformer) Transform(input string) string {
	return input
}

func Print(ma Matcher) string {
	m := ma.(*matcher)
	return m.tree.printTree(m.tree.root, 0)
}

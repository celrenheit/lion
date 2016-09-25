package matcher

type Matcher interface {
	Set(pattern string, values interface{}, tags Tags)
	Get(pattern string, tags Tags) (Context, interface{})
	GetWithContext(c Context, pattern string, tags Tags) interface{}
}

type GetSetterCreator interface {
	New() GetSetter
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
	GetSetterCreator GetSetterCreator
	ParamTransformer ParamTransformer
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

func (m *matcher) Set(pattern string, values interface{}, tags Tags) {
	m.tree.addRoute(m.tree.root, pattern, values, tags)
	m.postvalidation(pattern)
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
	return m.tree.getValue(n, tags)
}

func (m *matcher) postvalidation(pattern string) {
	// Find duplicate parameter names
	m.findDuplicateParamNames(m.tree.root, pattern, []string{})
}

func (m *matcher) findDuplicateParamNames(n *node, pattern string, pnames []string) {
	for _, children := range n.children {
		for _, child := range children {
			if child.nodeType > static && child.pname == "" {
				panicm(`cannot use an unnamed parameter for  %s`, pattern)
			}

			if len(child.pname) > 0 && isInStringSlice(pnames, child.pname) {
				panicm("lion: Duplicate parameter %s for %s", child.pname, pattern)
			}

			m.findDuplicateParamNames(child, pattern, append(pnames, child.pname))
		}
	}
}

type Tags []string

type noopParamTransformer struct{}

func (_ noopParamTransformer) Transform(input string) string {
	return input
}

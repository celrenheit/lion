package lion

type Module interface {
	Resource
	Base() string
	Routes(*Router)
}

type ModuleRequirements interface {
	Requires() []string
}

func (r *Router) Module(modules ...Module) {
	for _, m := range modules {
		r.registerModule(m)
	}
}

func (r *Router) registerModule(m Module) {
	g := r.Group(m.Base())
	if req, ok := m.(ModuleRequirements); ok {
		for _, dep := range req.Requires() {
			if !r.hasNamed(dep) {
				panic("Unmet middleware requirement for " + dep)
			}
			g.UseNamed(dep)
		}
	}

	g.Resource("/", m)

	m.Routes(g)
}

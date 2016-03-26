package lion

// Module represent an independent router entity.
// It should be used to group routes and subroutes together.
type Module interface {
	Resource
	Base() string
	Routes(*Router)
}

// ModuleRequirements specify that the module requires specific named middlewares.
type ModuleRequirements interface {
	Requires() []string
}

// Module register modules for the current router instance.
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

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

// ModuleAttacher attaches it self to the parent router. It gives the parent router as argument before grouping into the Base() routes.
// It useful for example to specify named middlewares that can be reused in the parent router.
type ModuleAttacher interface {
	Attach(*Router)
}

// Module register modules for the current router instance.
func (r *Router) Module(modules ...Module) {
	for _, m := range modules {
		r.registerModule(m)
	}
}

func (r *Router) registerModule(m Module) {
	if attacher, ok := m.(ModuleAttacher); ok {
		attacher.Attach(r)
	}

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

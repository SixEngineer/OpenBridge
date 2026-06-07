package tool

// 83

import "openbridge/backend/internal/domain/interfaces"

type Registry struct {
	providers map[string]interfaces.Provider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]interfaces.Provider),
	}
}

func (r *Registry) Register(name string, p interfaces.Provider) error {
	r.providers[name] = p
	return nil
}

func (r *Registry) Unregister(name string) {
	delete(r.providers, name)
}

func (r *Registry) Get(name string) (interfaces.Provider, bool) {
	p, exists := r.providers[name]
	return p, exists
}

func (r *Registry) MustGet(name string) interfaces.Provider {
	p, exists := r.providers[name]
	if !exists {
		panic("provider not found: " + name)
	}
	return p
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
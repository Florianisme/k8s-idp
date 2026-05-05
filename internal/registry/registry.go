package registry

import (
	"sort"
	"sync"
)

// Service represents a discovered service from a Pod or container.
type Service struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Owner         string `json:"owner,omitempty"`
	SourceURL     string `json:"sourceUrl,omitempty"`
	HasSpec       bool   `json:"hasSpec"`
	SpecPath      string `json:"-"`
	SpecPort      string `json:"-"`
	IP            string `json:"-"`
	Namespace     string `json:"namespace,omitempty"`
	PodName       string `json:"podName,omitempty"`
	ContainerName string `json:"containerName,omitempty"`
	ContainerID   string `json:"containerId,omitempty"`
}

// Registry is a thread-safe in-memory store of discovered services.
type Registry struct {
	mu       sync.RWMutex
	services map[string]Service
}

// New creates an empty Registry.
func New() *Registry {
	return &Registry{services: make(map[string]Service)}
}

// Add inserts or replaces a service by ID.
func (r *Registry) Add(s Service) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[s.ID] = s
}

// Remove deletes a service by ID. No-op if not present.
func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.services, id)
}

// Get returns a service by ID.
func (r *Registry) Get(id string) (Service, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.services[id]
	return s, ok
}

// List returns all services as a slice. Never returns nil.
func (r *Registry) List() []Service {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Service, 0, len(r.services))
	for _, s := range r.services {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

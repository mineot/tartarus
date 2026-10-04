package repositories

import "tartarus/store"

// Repos centralizes access to the shared store.
type Repos struct {
	Store *store.Store
}

// New creates a new Repos backed by the given store.
func New(s *store.Store) *Repos {
	return &Repos{Store: s}
}

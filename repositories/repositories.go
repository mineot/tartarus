package repositories

import "tartarus/store"

type Repos struct {
	Str *store.Store
}

func New(str *store.Store) *Repos {
	return &Repos{str}
}

// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package stdmap

import (
	"fmt"

	"github.com/dapperlabs/flow-go/model/flow"
)

// Collections implements a mempool storing collections.
type Collections struct {
	*Backend
}

// NewCollections creates a new memory pool for collection.
func NewCollections() (*Collections, error) {
	g := &Collections{
		Backend: NewBackend(),
	}

	return g, nil
}

// Add adds a collection to the mempool.
func (g *Collections) Add(coll *flow.Collection) error {
	return g.Backend.Add(coll)
}

// Get returns the collection with the given ID from the mempool.
func (g *Collections) Get(collID flow.Identifier) (*flow.Collection, error) {
	entity, err := g.Backend.Get(collID)
	if err != nil {
		return nil, err
	}
	coll, ok := entity.(*flow.Collection)
	if !ok {
		panic(fmt.Sprintf("invalid entity in collection pool (%T)", entity))
	}
	return coll, nil
}

// All returns all collections from the mempool.
func (g *Collections) All() []*flow.Collection {
	entities := g.Backend.All()
	colls := make([]*flow.Collection, 0, len(entities))
	for _, entity := range entities {
		coll, ok := entity.(*flow.Collection)
		if !ok {
			panic(fmt.Sprintf("invalid entity in collection pool (%T)", entity))
		}
		colls = append(colls, coll)
	}
	return colls
}

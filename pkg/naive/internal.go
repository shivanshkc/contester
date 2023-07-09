package naive

import (
	"contester/pkg/simulation"
	"sync"
)

type Internal struct {
	store      map[string]any
	storeMutex *sync.RWMutex
}

func NewInternal() *Internal {
	return &Internal{
		store:      map[string]any{},
		storeMutex: &sync.RWMutex{},
	}
}

// Set is a simple map set operation. It is thread-safe to use.
func (i *Internal) Set(ctx simulation.Context, key string, value any) (err error) {
	// Write lock.
	i.storeMutex.Lock()
	defer i.storeMutex.Unlock()

	if err := ctx.NetworkOp(); err != nil {
		return err
	}

	i.store[key] = value
	return nil
}

// Get is a simple map get operation. It is thread-safe to use.
func (i *Internal) Get(ctx simulation.Context, key string) (any, error) {
	// Read lock.
	i.storeMutex.RLock()
	defer i.storeMutex.RUnlock()

	if err := ctx.NetworkOp(); err != nil {
		return nil, err
	}

	return i.store[key], nil
}

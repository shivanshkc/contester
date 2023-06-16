package api

import (
	"fmt"
	"sync"
	"time"
)

// External represents the external API exposed by a node in the system.
type External interface {
	Get() (state string, err error)
	Set(state string) (err error)
}

// Internal represents a node's internal API that it only exposes to other nodes
// and not to the external user.
type Internal struct {
	// opFailureProbability is the probablity of an operation failing.
	// Its value should be between 0 and 1, both inclusive.
	opFailureProbability float64
	// opMinMaxDelayFn provides the minimum and maximum delay during
	// the execution of an operation.
	opMinMaxDelayFn func() (min, max time.Duration)

	store      map[string]any
	storeMutex *sync.RWMutex
}

// NewInternal creates a new Internal instance.
func NewInternal(opFailureProbability float64, opMinDelay, opMaxDelay time.Duration) *Internal {
	if opMinDelay > opMaxDelay {
		panic(fmt.Errorf("min delay is greater than max delay: min delay: %d, max delay %d", opMinDelay, opMaxDelay))
	}

	return &Internal{
		opFailureProbability: opFailureProbability,
		opMinMaxDelayFn:      func() (min, max time.Duration) { return opMinDelay, opMaxDelay },
		store:                map[string]any{},
		storeMutex:           &sync.RWMutex{},
	}
}

// Set is a simple map set operation. It is thread-safe to use.
func (i *Internal) Set(key string, value any) (err error) {
	// Write lock.
	i.storeMutex.Lock()
	defer i.storeMutex.Unlock()

	// Failing the operation artificially for the given probability.
	if biasedBoolean(i.opFailureProbability) {
		return fmt.Errorf("simulated error")
	}

	// Sleeping as per the given delay configs.
	time.Sleep(randomDurationBetween(i.opMinMaxDelayFn()))
	i.store[key] = value
	return nil
}

// Get is a simple map get operation. It is thread-safe to use.
func (i *Internal) Get(key string) (value any, err error) {
	// Read lock.
	i.storeMutex.RLock()
	defer i.storeMutex.RUnlock()

	// Failing the operation artificially for the given probability.
	if biasedBoolean(i.opFailureProbability) {
		return nil, fmt.Errorf("simulated error")
	}

	// Sleeping as per the given delay configs.
	time.Sleep(randomDurationBetween(i.opMinMaxDelayFn()))
	return i.store[key], nil
}

// UpdateArtificialFailureParams allows updating the artificial failure parameters of an existing Internal API.
// Note that it is NOT thread safe to use.
func (i *Internal) UpdateArtificialFailureParams(opFailureProbability float64, opMinDelay, opMaxDelay time.Duration) {
	i.opFailureProbability = opFailureProbability
	i.opMinMaxDelayFn = func() (min time.Duration, max time.Duration) { return opMinDelay, opMaxDelay }
}

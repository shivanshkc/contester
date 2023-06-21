package naive

import (
	"contester/pkg/utils"
	"fmt"
	"sync"
	"time"
)

type Internal struct {
	netFailureProbability float64
	networkPerformance    func() (min, max time.Duration)
	clockOffset           time.Duration

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
func (i *Internal) Set(key string, value any) (err error) {
	// Write lock.
	i.storeMutex.Lock()
	defer i.storeMutex.Unlock()

	// Failing the operation artificially for the given probability.
	if utils.BiasedBoolean(i.netFailureProbability) {
		return fmt.Errorf("simulated error")
	}

	// Sleeping as per the given delay configs.
	time.Sleep(utils.RandomDurationBetween(i.networkPerformance()))
	i.store[key] = value
	return nil
}

// Get is a simple map get operation. It is thread-safe to use.
func (i *Internal) Get(key string) (value any, err error) {
	// Read lock.
	i.storeMutex.RLock()
	defer i.storeMutex.RUnlock()

	// Failing the operation artificially for the given probability.
	if utils.BiasedBoolean(i.netFailureProbability) {
		return nil, fmt.Errorf("simulated error")
	}

	// Sleeping as per the given delay configs.
	time.Sleep(utils.RandomDurationBetween(i.networkPerformance()))
	return i.store[key], nil
}

func (i *Internal) SetNetworkFailureProbability(probability float64) {
	if probability < 0 || probability > 1 {
		panic("probability should be in the interval [0, 1]")
	}

	i.netFailureProbability = probability
}

func (i *Internal) SetNetworkPerformance(minDelay, maxDelay time.Duration) {
	if maxDelay < minDelay {
		panic("maxDelay should be >= minDelay")
	}
	if minDelay < 0 {
		panic("delay parameters should be positive")
	}

	i.networkPerformance = func() (min time.Duration, max time.Duration) { return minDelay, maxDelay }
}

func (i *Internal) SetClockOffset(offset time.Duration) {
	i.clockOffset = offset
}

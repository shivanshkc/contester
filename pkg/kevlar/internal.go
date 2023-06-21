package kevlar

import (
	"contester/pkg/utils"
	"errors"
	"fmt"
	"sync"
	"time"
)

const lockTimeout = time.Minute

type lockInfo struct {
	LockID    string
	ExpiresAt time.Time
}

// record represents the data structure used by Kevlar to store a key-value pair.
type record struct {
	// Key is the identifier of this Record.
	Key string

	// ConfirmedValue is the most recent, confirmed as successfully written, value of this Record.
	ConfirmedValue string
	// UnconfirmedValue is the most recently written value of this Record. It is not confirmed as successfully written.
	UnconfirmedValue string

	// Version is number of times the ConfirmedValue has been updated for this Record.
	Version int64
	// Signature of the request that last updated this record.
	Signature string
}

type Internal struct {
	netFailureProbability float64
	networkPerformance    func() (min, max time.Duration)
	clockOffset           time.Duration

	store      map[string]*record
	storeMutex *sync.RWMutex
	keyLockMap map[string]*lockInfo
}

func NewInternal() *Internal {
	return &Internal{
		store:      map[string]*record{},
		storeMutex: &sync.RWMutex{},
		keyLockMap: map[string]*lockInfo{},
	}
}

func (i *Internal) get(key string) (*record, error) {
	// Read lock.
	i.storeMutex.RLock()
	defer i.storeMutex.RUnlock()

	// Failing the operation artificially for the given probability.
	if utils.BiasedBoolean(i.netFailureProbability) {
		return nil, fmt.Errorf("simulated error")
	}

	// Sleeping as per the given delay configs.
	time.Sleep(utils.RandomDurationBetween(i.networkPerformance()))

	rec, exists := i.store[key]
	if !exists {
		rec = &record{Key: key, Version: -1}
	}

	return rec, nil
}

func (i *Internal) getAndLock(key string, lockID string) (*record, error) {
	// Write lock because of a potential write operation.
	i.storeMutex.Lock()
	defer i.storeMutex.Unlock()

	// Failing the operation artificially for the given probability.
	if utils.BiasedBoolean(i.netFailureProbability) {
		return nil, fmt.Errorf("simulated error")
	}

	// Sleeping as per the given delay configs.
	time.Sleep(utils.RandomDurationBetween(i.networkPerformance()))

	lock, exists := i.keyLockMap[key]
	// If lock exists and is not expired...
	if exists && !time.Now().After(lock.ExpiresAt) {
		return nil, errors.New("key already locked")
	}

	// Create a new lock record.
	i.keyLockMap[key] = &lockInfo{
		LockID:    lockID,
		ExpiresAt: time.Now().Add(lockTimeout),
	}

	rec, exists := i.store[key]
	if !exists {
		rec = &record{Key: key, Version: -1}
	}

	return rec, nil
}

func (i *Internal) setAndUnlock(key string, value *record, lockID string) error {
	// Write lock because of a potential write operation.
	i.storeMutex.Lock()
	defer i.storeMutex.Unlock()

	// Failing the operation artificially for the given probability.
	if utils.BiasedBoolean(i.netFailureProbability) {
		return fmt.Errorf("simulated error")
	}

	// Sleeping as per the given delay configs.
	time.Sleep(utils.RandomDurationBetween(i.networkPerformance()))

	lock, exists := i.keyLockMap[key]
	// If lock does not exist or is expired...
	if !exists || time.Now().After(lock.ExpiresAt) {
		return errors.New("key not locked")
	}
	if lock.LockID != lockID {
		return errors.New("lock ID does not match")
	}

	// Unlock the key.
	delete(i.keyLockMap, key)
	// Store value.
	i.store[key] = value

	return nil
}

func (i *Internal) unlock(key string, lockID string) error {
	// Write lock because of a potential write operation.
	i.storeMutex.Lock()
	defer i.storeMutex.Unlock()

	// Failing the operation artificially for the given probability.
	if utils.BiasedBoolean(i.netFailureProbability) {
		return fmt.Errorf("simulated error")
	}

	// Sleeping as per the given delay configs.
	time.Sleep(utils.RandomDurationBetween(i.networkPerformance()))

	lock, exists := i.keyLockMap[key]
	if !exists {
		return nil
	}
	if lock.LockID != lockID {
		return errors.New("lock ID does not match")
	}

	// Unlock the key.
	delete(i.keyLockMap, key)
	return nil
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

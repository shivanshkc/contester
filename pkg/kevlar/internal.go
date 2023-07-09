package kevlar

import (
	"contester/pkg/simulation"
	"errors"
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

func (i *Internal) get(ctx simulation.Context, key string) (*record, error) {
	// Read lock.
	i.storeMutex.RLock()
	defer i.storeMutex.RUnlock()

	if err := ctx.NetworkOp(); err != nil {
		return nil, err
	}

	rec, exists := i.store[key]
	if !exists {
		rec = &record{Key: key, Version: -1}
	}

	return rec, nil
}

func (i *Internal) getAndLock(ctx simulation.Context, key string, lockID string) (*record, error) {
	// Write lock because of a potential write operation.
	i.storeMutex.Lock()
	defer i.storeMutex.Unlock()

	if err := ctx.NetworkOp(); err != nil {
		return nil, err
	}

	lock, exists := i.keyLockMap[key]
	// If lock exists and is not expired...
	if exists && !ctx.Time().After(lock.ExpiresAt) {
		return nil, errors.New("key already locked")
	}

	// Create a new lock record.
	i.keyLockMap[key] = &lockInfo{
		LockID:    lockID,
		ExpiresAt: ctx.Time().Add(lockTimeout),
	}

	rec, exists := i.store[key]
	if !exists {
		rec = &record{Key: key, Version: -1}
	}

	return rec, nil
}

func (i *Internal) setAndUnlock(ctx simulation.Context, key string, value *record, lockID string) error {
	// Write lock because of a potential write operation.
	i.storeMutex.Lock()
	defer i.storeMutex.Unlock()

	if err := ctx.NetworkOp(); err != nil {
		return err
	}

	lock, exists := i.keyLockMap[key]
	// If lock does not exist or is expired...
	if !exists || ctx.Time().After(lock.ExpiresAt) {
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

func (i *Internal) unlock(ctx simulation.Context, key string, lockID string) error {
	// Write lock because of a potential write operation.
	i.storeMutex.Lock()
	defer i.storeMutex.Unlock()

	if err := ctx.NetworkOp(); err != nil {
		return err
	}

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

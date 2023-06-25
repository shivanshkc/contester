package kevlar

import (
	"contester/pkg/simulation"
	"contester/pkg/utils"
	"errors"
	"math"

	"github.com/google/uuid"
)

// External implements the simulation.External interface using Kevlar.
// Read the method descriptions to understand the algorithm.
//
// Note that this implementation guarantees consensus.
type External struct {
	InternalAPIs []*Internal
}

func NewExternal(internalAPIs []*Internal) *External {
	return &External{InternalAPIs: internalAPIs}
}

// Get TODO
func (e *External) Get(ctx simulation.Context) (string, error) {
	// Get state values from all keepers.
	records, errs := e.getStateFromAll(ctx)
	// If majority failed, end execution right away.
	if len(errs) >= utils.GetSmallestMajority(len(e.InternalAPIs)) {
		return "", errors.Join(errs...)
	}

	// Determining the status of the last write request. It is equivalent to defining the state of the system.
	lws, state := e.determineLWS(records)
	// Taking action based on the LWS.
	switch lws {
	case "success":
		return state.UnconfirmedValue, nil
	case "failure":
		return state.ConfirmedValue, nil
	case "unknown":
		return "", errors.Join(errs...)
	default:
		return "", errors.Join(errs...)
	}

}

// Set TODO
func (e *External) Set(ctx simulation.Context, state string) error {
	// Generate a new lockID.
	lockID := uuid.NewString()
	smMajority := utils.GetSmallestMajority(len(e.InternalAPIs))

	// Get state from all keepers and lock them for writing.
	records, errs := e.getAndLockStateFromAll(ctx, lockID)
	// Unlock all keepers at the end, even if setAndUnlock passes, for safety.
	defer func() { _ = e.unlockAll(ctx, lockID) }()

	// If majority failed, end execution.
	if len(errs) >= smMajority {
		return errors.Join(errs...)
	}

	// This record will be set on all the State-Keepers.
	newState := &record{Key: "state"}

	// Determining the status of the last write request. It is equivalent to defining the state of the system.
	lws, currentState := e.determineLWS(records)
	// Taking action based on the LWS.
	switch lws {
	// If the last write request was successful, we upgrade its value to the confirmed value, and increase the version
	// of the record. Note that this version increase is solely due to the success of the last request and has nothing
	// to do with this request.
	case "success":
		// Form the new state.
		newState.ConfirmedValue = currentState.UnconfirmedValue
		newState.UnconfirmedValue = state
		newState.Version = currentState.Version + 1
		newState.Signature = uuid.NewString()
	// If the last write request was unsuccessful, we retain the last confirmed value and version.
	case "failure":
		// Form the new state.
		newState.ConfirmedValue = currentState.ConfirmedValue
		newState.UnconfirmedValue = state
		newState.Version = currentState.Version
		newState.Signature = uuid.NewString()
	case "unknown":
		return errors.Join(errs...)
	default:
		return errors.Join(errs...)
	}

	// Setting the new state in all State-Keepers.
	errSet := e.setAndUnlockStateOnAll(ctx, newState, lockID)
	// If a majority of keepers reject, we consider the operation failed.
	if len(errSet) >= smMajority {
		return errors.Join(errSet...)
	}

	// Value successfully written. It is now guaranteed to be promoted to the ConfirmedValue eventually.
	return nil
}

// determineLWS stands for determine-last-write-status.
//
// It uses the provided list of records to determine the status of the most recent write request(s).
// This method constitutes the most of Kevlar.
//
// Returned parameters:
//  1. string - Status of the last write request. It can be "success", "failure" or "unknown".
//  2. *Record - Record that holds the current state of the system. It will be nil if the state is "unknown".
func (e *External) determineLWS(records []*record) (string, *record) {
	// Get the smallest majority number.
	smMajority := utils.GetSmallestMajority(len(e.InternalAPIs))

	// Determining the highest version amongst all records.
	var highestVersion int64 = math.MinInt64
	for _, rec := range records {
		if highestVersion < rec.Version {
			highestVersion = rec.Version
		}
	}

	// This will map signatures to the number of records that they belong to.
	signatureCounts := map[string]int{}
	// This will hold the biggest signature count.
	biggestSignatureCount := math.MinInt64
	// This will hold any record that has the highest version.
	var anyHighestVersionRecord *record

	// Looping to populate the above quantity.
	for _, rec := range records {
		// If the record is stale, it can be ignored.
		if rec.Version < highestVersion {
			continue
		}

		// Updating the map.
		signatureCounts[rec.Signature]++
		// If any one map element contains a majority of records, we have determined the state.
		if signatureCounts[rec.Signature] >= smMajority {
			return "success", rec
		}

		// Updating the biggestSignatureCount.
		if thisCount := signatureCounts[rec.Signature]; biggestSignatureCount < thisCount {
			biggestSignatureCount = thisCount
		}

		// This will be required if the control goes below this loop.
		anyHighestVersionRecord = rec
	}

	// This is the number of nodes that are either down or failed to communicate.
	deadKeeperCount := len(e.InternalAPIs) - len(records)

	// If this is true, we can be certain that the most recent requests have all failed.
	if biggestSignatureCount+deadKeeperCount < smMajority {
		return "failure", anyHighestVersionRecord
	}

	// The dead keepers may contain data which is necessary to determine the state.
	// So, until those keepers come up, the state cannot be determined.
	return "unknown", nil
}

// getStateFromAll gets the records from all keepers concurrently.
func (e *External) getStateFromAll(ctx simulation.Context) ([]*record, []error) {
	// This channel will store the result of the internal API calls.
	respChan := make(chan func() (*record, error), len(e.InternalAPIs))
	defer close(respChan)

	// Looping over all internal APIs and getting the state from them all.
	for _, iAPI := range e.InternalAPIs {
		go func(iAPI *Internal) {
			value, err := iAPI.get(ctx, "state")
			respChan <- func() (*record, error) { return value, err }
		}(iAPI)
	}

	var errs []error
	var records []*record

	// Looping again to collect results.
	for i := 0; i < len(e.InternalAPIs); i++ {
		value, err := (<-respChan)()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		records = append(records, value)
	}

	return records, errs
}

// getAndLockStateFromAll gets records from all keepers concurrently and locks them for writing.
func (e *External) getAndLockStateFromAll(ctx simulation.Context, lockID string) ([]*record, []error) {
	// This channel will store the result of the internal API calls.
	respChan := make(chan func() (*record, error), len(e.InternalAPIs))
	defer close(respChan)

	// Looping over all internal APIs and getting the state from them all.
	for _, iAPI := range e.InternalAPIs {
		go func(iAPI *Internal) {
			value, err := iAPI.getAndLock(ctx, "state", lockID)
			respChan <- func() (*record, error) { return value, err }
		}(iAPI)
	}

	var errs []error
	var records []*record

	// Looping again to collect results.
	for i := 0; i < len(e.InternalAPIs); i++ {
		value, err := (<-respChan)()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		records = append(records, value)
	}

	return records, errs
}

// setAndUnlockStateOnAll sets the given record in all keepers and unlocks them for writing.
func (e *External) setAndUnlockStateOnAll(ctx simulation.Context, rec *record, lockID string) []error {
	// This channel will store the result of the internal API calls.
	respChan := make(chan error, len(e.InternalAPIs))
	defer close(respChan)

	// Looping over all internal APIs and getting the state from them all.
	for i, iAPI := range e.InternalAPIs {
		go func(i int, iAPI *Internal) {
			respChan <- iAPI.setAndUnlock(ctx, "state", rec, lockID)
		}(i, iAPI)
	}

	var errs []error

	// Looping again to collect results.
	for i := 0; i < len(e.InternalAPIs); i++ {
		err := <-respChan
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// unlockAll unlocks all keepers.
func (e *External) unlockAll(ctx simulation.Context, lockID string) []error {
	// This channel will store the result of the internal API calls.
	respChan := make(chan error, len(e.InternalAPIs))
	defer close(respChan)

	// Looping over all internal APIs and getting the state from them all.
	for i, iAPI := range e.InternalAPIs {
		go func(i int, iAPI *Internal) {
			respChan <- iAPI.unlock(ctx, "state", lockID)
		}(i, iAPI)
	}

	var errs []error

	// Looping again to collect results.
	for i := 0; i < len(e.InternalAPIs); i++ {
		err := <-respChan
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

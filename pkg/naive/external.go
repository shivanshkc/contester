package naive

import (
	"contester/pkg/simulation"
	"contester/pkg/utils"
	"errors"
	"fmt"
)

// External implements the simulation.ExternalAPI interface using the naive majority approach.
// Read the method descriptions to understand the algorithm.
//
// Note that this implementation does NOT guarantee consensus.
type External struct {
	InternalAPIs []*Internal
}

func NewExternal(internalAPIs []*Internal) *External {
	return &External{InternalAPIs: internalAPIs}
}

// Get collects the state from all the internal APIs.
// If a majority of calls fail, the operation is considered failed.
// Otherwise, if a single value exists on a majority of nodes, it is considered valid state and returned.
// Otherwise, a consensus error is returned.
func (e *External) Get(ctx simulation.Context) (string, error) {
	// This channel will store the result of the internal API calls.
	respChan := make(chan func() (any, error), len(e.InternalAPIs))
	defer close(respChan)

	// Looping over all internal APIs and getting the state from them all.
	for _, iAPI := range e.InternalAPIs {
		go func(iAPI *Internal) {
			value, err := iAPI.Get(ctx, "state")
			respChan <- func() (any, error) { return value, err }
		}(iAPI)
	}

	// This slice will collect errors.
	// If they are in majority, the operation will be considered failed
	// and the joined error will be returned.
	var errs []error
	// This will map different state values to their occurrence count.
	// We need a single value to exist in majority for it to be the agreed upon state value.
	valueCounts := map[string]int{}

	// Looping again to collect results.
	for i := 0; i < len(e.InternalAPIs); i++ {
		value, err := (<-respChan)()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// Just a type safety check to be sure.
		valueStr, ok := value.(string)
		if !ok {
			panic("value is not a string, this is unexpected")
		}

		// Increase the value occurrence count.
		valueCounts[valueStr]++
	}

	// Get the smallest majority number.
	smMajority := utils.GetSmallestMajority(len(e.InternalAPIs))
	// If a majority of calls failed, the operation is failed.
	if len(errs) >= smMajority {
		return "", errors.Join(errs...)
	}

	// Looping over all values to see if any exists on a majority.
	for value, count := range valueCounts {
		if count >= smMajority {
			return value, nil
		}
	}

	// No value exists on a majority, the algorithm has failed.
	return "", fmt.Errorf("no consensus on any value, values: %+v", valueCounts)
}

// Set sets the state on all the internal APIs.
// If a majority of calls fail, the operation is considered failed.
// Otherwise, the operation is considered successful.
func (e *External) Set(ctx simulation.Context, state string) error {
	// This channel will store the result of the internal API calls.
	respChan := make(chan error, len(e.InternalAPIs))
	defer close(respChan)

	// Looping over all internal APIs and setting the state on them all.
	for _, iAPI := range e.InternalAPIs {
		go func(iAPI *Internal) {
			respChan <- iAPI.Set(ctx, "state", state)
		}(iAPI)
	}

	// This slice will collect errors.
	// If they are in majority, the operation will be considered failed
	// and the joined error will be returned.
	var errs []error

	// Looping again to collect results.
	for i := 0; i < len(e.InternalAPIs); i++ {
		if err := <-respChan; err != nil {
			errs = append(errs, err)
		}
	}

	// Get the smallest majority number.
	smMajority := utils.GetSmallestMajority(len(e.InternalAPIs))
	// If a majority of calls failed, the operation is failed.
	if len(errs) >= smMajority {
		return errors.Join(errs...)
	}

	// The operation was a success.
	return nil
}

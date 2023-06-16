package api_test

import (
	"consim/pkg/api"
	"time"

	"github.com/goombaio/namegenerator"
)

// _nameGen generates random readable values.
var _nameGen = namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())

// getRandomValue generates a random string.
func getRandomValue() string {
	return _nameGen.Generate()
}

// createAPIs creates and returns the given number of External and Internal API instances.
// It adds simulation parameters to the InternalAPIs and attaches them to the ExternalAPIs.
func createAPIs(nodeCount int) ([]api.External, []*api.Internal) {
	external, internal := make([]api.External, nodeCount), make([]*api.Internal, nodeCount)

	// Create internal APIs.
	for i := 0; i < nodeCount; i++ {
		internal[i] = api.NewInternal(0.25, time.Microsecond, 100*time.Microsecond)
	}

	// Create external APIs.
	for i := 0; i < nodeCount; i++ {
		external[i] = &api.ExternalSimpleMajority{InternalAPIs: internal}
	}

	return external, internal
}

// noArtificialFailures updates all the given Internal APIs so that there
// are no artificial failures or delays in their working.
func noArtificialFailures(inAPIs []*api.Internal) {
	for _, inAPI := range inAPIs {
		inAPI.UpdateArtificialFailureParams(0, 0, 0)
	}
}

package simulation_test

import (
	"contester/pkg/kevlar"
	"contester/pkg/naive"
	"contester/pkg/simulation"
	"time"

	"github.com/goombaio/namegenerator"
)

// kevlarAPIs creates and returns the given number of External and Internal API instances of Kevlar.
func kevlarAPIs(nodeCount int) ([]simulation.External, []simulation.Internal) {
	external, internal := make([]simulation.External, nodeCount), make([]simulation.Internal, nodeCount)
	internalConcrete := make([]*kevlar.Internal, nodeCount)

	// Create internal APIs.
	for i := 0; i < nodeCount; i++ {
		inter := kevlar.NewInternal()
		inter.SetNetworkFailureProbability(0.1)
		inter.SetNetworkPerformance(time.Microsecond, time.Millisecond)
		inter.SetClockOffset(time.Millisecond)

		internal[i], internalConcrete[i] = inter, inter
	}

	// Create external APIs.
	for i := 0; i < nodeCount; i++ {
		external[i] = kevlar.NewExternal(internalConcrete)
	}

	return external, internal
}

// naiveAPIs creates and returns the given number of External and Internal API instances of Naive.
func naiveAPIs(nodeCount int) ([]simulation.External, []simulation.Internal) {
	external, internal := make([]simulation.External, nodeCount), make([]simulation.Internal, nodeCount)
	internalConcrete := make([]*naive.Internal, nodeCount)

	// Create internal APIs.
	for i := 0; i < nodeCount; i++ {
		inter := naive.NewInternal()
		inter.SetNetworkFailureProbability(0.1)
		inter.SetNetworkPerformance(time.Microsecond, time.Millisecond)
		inter.SetClockOffset(time.Millisecond)

		internal[i], internalConcrete[i] = inter, inter
	}

	// Create external APIs.
	for i := 0; i < nodeCount; i++ {
		external[i] = naive.NewExternal(internalConcrete)
	}

	return external, internal
}

// _nameGen generates random readable values.
var _nameGen = namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())

// getRandomValue generates a random string.
func getRandomValue() string {
	return _nameGen.Generate()
}

// noArtificialFailures updates all the given Internal APIs so that there
// are no artificial failures or delays in their working.
func noArtificialFailures(inAPIs []simulation.Internal) {
	for _, inAPI := range inAPIs {
		inAPI.SetNetworkFailureProbability(0)
		inAPI.SetNetworkPerformance(0, 0)
		inAPI.SetClockOffset(0)
	}
}

package main

import (
	"fmt"

	"contester/pkg/kevlar"
	"contester/pkg/naive"
	"contester/pkg/simulation"
)

const (
	// Number of nodes in the system to be tested.
	// Feel free to change it to any number > 1
	nodeCount = 5

	// Number of times the simulation should run.
	runCount = 100
)

func main() {
	for i := 0; i < runCount; i++ {
		// Create new instances for every run.
		instances := createKevlarInstances()

		// Run the simulation with newly created instances.
		err := simulation.Run(simulation.QuickStartConfig, instances)
		if err != nil {
			panic(err)
		}

		fmt.Printf("\rSession %d/%d passed.", i+1, runCount)
	}

	fmt.Println("\nConsensus maintained.")
}

func createNaiveInstances() []simulation.ExternalAPI {
	internalAPIs := make([]*naive.Internal, nodeCount)
	externalAPIs := make([]simulation.ExternalAPI, nodeCount)

	for i := 0; i < nodeCount; i++ {
		internalAPIs[i] = naive.NewInternal()
	}
	for i := 0; i < nodeCount; i++ {
		externalAPIs[i] = naive.NewExternal(internalAPIs)
	}

	return externalAPIs
}

func createKevlarInstances() []simulation.ExternalAPI {
	internalAPIs := make([]*kevlar.Internal, nodeCount)
	externalAPIs := make([]simulation.ExternalAPI, nodeCount)

	for i := 0; i < nodeCount; i++ {
		internalAPIs[i] = kevlar.NewInternal()
	}
	for i := 0; i < nodeCount; i++ {
		externalAPIs[i] = kevlar.NewExternal(internalAPIs)
	}

	return externalAPIs
}

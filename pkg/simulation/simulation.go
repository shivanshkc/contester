package simulation

import (
	"context"
	"fmt"
	"time"
)

// ExternalAPI represents an API that a node in a distributed system
// exposes to the outside world.
//
// The simulation tests if this API exhibits CP behaviour when invoked
// in parallel across multiple nodes in the system.
//
// Importantly, implementations should-
//  1. Use ctx.Time() method instead of time.Now() function to get the
//     current time.
//  2. Call ctx.NetworkOp() method before an IPC which would've been a
//     network operation in an actual system.
//
// The calls listed above make sure that the implementation respects the
// simulation configs.
type ExternalAPI interface {
	// Get provides the current state of the system and an error if
	// something goes wrong.
	Get(ctx Context) (state string, err error)
	// Set sets the state of the system, and returns an error if
	// something goes wrong.
	Set(ctx Context, state string) (err error)
}

// Run the simulation for the given configs and node instances.
func Run(conf Config, instances []ExternalAPI) error {
	// Validate the user provided config.
	if err := conf.validate(); err != nil {
		return fmt.Errorf("invalid config provided: %w", err)
	}

	// Create context for the simulation.
	simulationCtx := kontext{
		Context: context.Background(),
		conf:    conf,
	}

	// Run the simulation with all validated parameters.
	if err := run(simulationCtx, instances); err != nil {
		return err // No wrapping required.
	}

	// Consensus maintained.
	return nil
}

// run a simulation session.
func run(ctx kontext, instances []ExternalAPI) error {
	// Send the required number of requests.
	states := sendRoundRobinRequests(ctx, instances)

	// Determine the expected state.
	var expectedState string
	if len(states) != 0 {
		expectedState = states[len(states)-1]
	}

	// Use ideal config for getting the current state.
	ctx.conf = idealConfig
	// Get the current/actual state.
	actualState, err := instances[0].Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	// Verify the state.
	if actualState != expectedState {
		return fmt.Errorf("consensus broken. expected state: %s, but got: %s",
			expectedState, actualState)
	}

	return nil
}

// sendRoundRobinRequests sends the configured number of requests in round-robin
// fashion to the provided instances.
//
// It ignores failures, and returns the list of successful resposes, in the SAME
// order as they were received.
func sendRoundRobinRequests(ctx kontext, instances []ExternalAPI) []string {
	// Short hand for config.
	conf := ctx.conf

	// The channel that will receive all responses from the system.
	responseChan := make(chan func() (string, error), conf.RequestCount)
	defer close(responseChan)

	// Get node count for easy usage below.
	nodeCount := int64(len(instances))

	// Call the external API in round-robin requestCount-times.
	for i := int64(0); i < conf.RequestCount; i++ {
		go func(i int64) {
			// Generate a random state for every request.
			state := getRandomValue()
			// External API call.
			err := instances[i%nodeCount].Set(ctx, state)
			responseChan <- func() (string, error) { return state, err }
		}(i)

		// Sleep for some time before sending another request.
		// This avoids "true simultaneity".
		time.Sleep(conf.RequestInterval)
	}

	var success []string
	// Collect all non-error responses.
	for i := int64(0); i < conf.RequestCount; i++ {
		state, err := (<-responseChan)()
		if err == nil {
			success = append(success, state)
		}
	}

	return success
}

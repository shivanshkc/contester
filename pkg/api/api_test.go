package api_test

import (
	"testing"
	"time"
)

func TestConsensusBulk(t *testing.T) {
	for i := 0; i < 100; i++ {
		if TestConsensusSingle(t); t.Failed() {
			return
		}
	}
}

func TestConsensusSingle(t *testing.T) {
	// Some simulation configs.
	const requestCount, nodeCount, requestInterval = 1000, 5, time.Microsecond
	// Create API instances.
	externalAPIs, internalAPIs := createAPIs(nodeCount)

	// The channel that will receive all responses from the system.
	responseChan := make(chan func() (string, error), requestCount)
	defer close(responseChan)

	// Call the external API in round robin requestCount-times.
	for i := 0; i < requestCount; i++ {
		go func(i int) {
			// This is the nodeIndex that ensures round-robin invocations.
			nodeIdxToUse := i - int(i/nodeCount)*nodeCount
			// Generate a random state for every request.
			state := getRandomValue()
			// System call.
			err := externalAPIs[nodeIdxToUse].Set(state)
			responseChan <- func() (string, error) { return state, err }
		}(i)

		// Sleep for some time before sending another request.
		// This avoids "true simultaneity".
		time.Sleep(requestInterval)
	}

	// This will hold the last acknowledged state.
	var expectedState string
	// This loop will update the expectedState var.
	for i := 0; i < requestCount; i++ {
		state, err := (<-responseChan)()
		if err == nil {
			expectedState = state
		}
	}

	// Remove all failure probabilities.
	noArtificialFailures(internalAPIs)

	// Fetch the current state of the system.
	currentState, err := externalAPIs[0].Get()
	if err != nil {
		t.Errorf("expected no error but got: %s", err.Error())
		return
	}

	// Verify the state.
	if currentState != expectedState {
		t.Errorf("expected final state to be: %s, but got: %s", expectedState, currentState)
	}

	t.Logf("Consensus maintained. Expected state = Final state = %s", currentState)
}

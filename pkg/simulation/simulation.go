package simulation

import (
	"time"
)

// External represents the external API exposed by a node in the system.
type External interface {
	// Get returns the current agreed upon state of the system (ideally).
	Get() (string, error)
	// Set sets the state of the system.
	Set(string) error

	// IdealOperation sets all failure probabilities to zero.
	IdealOperation()
}

// Internal represents a mock internal API of node in the system.
// It provides various methods to simulate the challenges of an actual
// distributed system.
type Internal interface {
	// SetNetworkFailureProbability sets the network failure probability of
	// the implementation.
	SetNetworkFailureProbability(float64)
	// SetNetworkPerformance sets the minimum and maximum delay for any
	// network operation (which in this case is just IPC).
	SetNetworkPerformance(minDelay, maxDelay time.Duration)
	// SetClockOffset adds an offset to the clock of the implementation.
	// This is because, practically, clocks on different machines are always
	// out of sync.
	SetClockOffset(time.Duration)
}

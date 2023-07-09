package simulation

import (
	"fmt"
	"time"
)

// idealConfig is the config for an ideal simulation where there
// are no network faults, and different clocks never go out of sync.
var idealConfig = Config{
	RequestCount:              0, // NO NEED TO SET.
	RequestInterval:           0, // NO NEED TO SET.
	NetworkFailureProbability: 0, // Perfectly stable networks.
	NetworkMinDelay:           0, // Infinite speed.
	NetworkMaxDelay:           0, // Infinite speed.
	MaxClockOffset:            0, // Perfectly synced clocks.
}

// QuickStartConfig to get started with a simulation.
var QuickStartConfig = Config{
	RequestCount:              10,
	RequestInterval:           time.Microsecond,
	NetworkFailureProbability: 0.1,
	NetworkMinDelay:           time.Millisecond / 10,
	NetworkMaxDelay:           time.Millisecond,
	MaxClockOffset:            10 * time.Millisecond,
}

// Config for the simulation.
type Config struct {
	// RequestCount is the total number of requests
	// that will be fed to the system.
	RequestCount int64
	// RequestInterval is the delay between two consecutive requests.
	//
	// This should be set, otherwise the delay between two consecutive
	// requests could be in nanoseconds (since it is all IPC), which
	// would make their true order difficult to detect.
	RequestInterval time.Duration
	// NetworkFailureProbability is a number in the interval [0, 1]
	// and represents the failure probability of a network operation.
	NetworkFailureProbability float64
	// NetworkMinDelay is the minimum delay of a network operation.
	NetworkMinDelay time.Duration
	// NetworkMaxDelay is the maximum delay of a network operation.
	NetworkMaxDelay time.Duration
	// MaxClockOffset is the maximum offset a clock can have in the
	// simulation, as no two systems have perfectly synced clocks.
	MaxClockOffset time.Duration
}

// validate the user provided config.
func (c Config) validate() error {
	if c.RequestCount < 2 {
		return fmt.Errorf("request count must be at least 2")
	}

	if c.RequestInterval < 0 {
		return fmt.Errorf("request interval must be > 0")
	}

	if c.NetworkFailureProbability < 0 || c.NetworkFailureProbability > 1 {
		return fmt.Errorf("network failure probability must be in the interval [0, 1]")
	}

	if c.NetworkMinDelay > c.NetworkMaxDelay {
		return fmt.Errorf("network min delay must be <= network max delay")
	}

	if c.NetworkMaxDelay < 0 {
		return fmt.Errorf("network delays cannot be negative")
	}

	if c.MaxClockOffset < 0 {
		return fmt.Errorf("max clock offset cannot be negative")
	}

	return nil
}

package simulation

import (
	"context"
	"fmt"
	"time"
)

// Context encapsulates the standard Go context along with methods
// that maybe required by the ExternalAPI implementations to be
// simulated correctly.
type Context interface {
	context.Context

	// NetworkOp is a dummy network operation that follows
	// the artificial network rules of the simulation.
	//
	// An ExternalAPI implementation should call this method
	// before an IPC which would've been a network operation
	// in an actual system.
	NetworkOp() error

	// Time provides the current time that is offset as per
	// the simulaton's configs.
	Time() time.Time
}

// kontext implements the Context interface.
type kontext struct {
	context.Context

	conf Config
}

func (k kontext) NetworkOp() error {
	// Fail the operation artificially for the given probability.
	if biasedBoolean(k.conf.NetworkFailureProbability) {
		return fmt.Errorf("artificial network failure")
	}

	// Sleep as per the given delay configs.
	time.Sleep(randomDurationBetween(k.conf.NetworkMinDelay, k.conf.NetworkMaxDelay))
	return nil
}

func (k kontext) Time() time.Time {
	// Add the specified offset to the current time.
	return time.Now().Add(randomDurationBetween(0, k.conf.MaxClockOffset))
}

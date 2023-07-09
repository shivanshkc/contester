# Contester

A program that tests any consensus algorithm implementation by running it in simulated network partitions and faults.

## How to use
The simulation runs when the `simulation.Run` function is called. It accepts `simulation.Config` and a slice of `simulation.ExternalAPI` interfaces.

For the first parameter, the simulation package provides a quickstart config, called `simulation.QuickStartConfig`. Users can provide their custom simulation config as well. Go through the code comments on the `simulation.Config` struct to understand the meaning of all fields.

The second parameter requires you to implement the consensus algorithm that needs to be tested. The type of the parameter is `[]simulation.ExternalAPI`. Here, each element of the slice represents a node in the distributed system.

The methods of the `simulation.ExternalAPI` type are invoked with `simulation.Context` instead of Go's standard `context.Context`. This is because the custom context type encapsulates methods that should be used by the implementations to be simulated correctly.

To learn more about how to write a `simulation.ExternalAPI` implementation, go through the existing implementations, namely `pkg/kevlar` and `pkg/naive`.
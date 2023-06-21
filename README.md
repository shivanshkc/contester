# Contester

A program that tests any consensus algorithm implementation by running it in simulated network partitions and faults.

## How to use
Every node in a distributed system has two interfaces. One that it exposes to external clients, and one that it exposes to other nodes in the system.
The `pkg/simulation` package declares two interfaces, `External` and `Internal` that mimic the same behaviour.

In order to test a new consensus algorithm-  
1. Implement the `simulation.External` and `simulation.Internal` interfaces with the consensus logic of the algorithm.
1. Write a function that creates instances of your implementation. See the `naiveAPIs` function in the `pkg/simulation/utils_test.go` file.
1. Replace the current instance creation call with the one you just wrote in the initial lines of the `TestConsensusSingle` function in the `pkg/simulation/simulation_test.go` file.
1. Run `go test -timeout 600s -run ^TestConsensusBulk$ contester/pkg/simulation -v`
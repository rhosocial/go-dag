# Package simple

## Features

- Workflow:
  - No limit to the complexity of a directed acyclic graph.
  - Supports single input and single output.
  - Can be canceled at any time during execution.
  - Comes with support for various loggers.
- Transit:
  - Provides `TransitInterface` for custom transits.
  - Allows any inputs and outputs (excluding initial input and final output).
  - Workers accept the incoming context parameter and can receive a `Done()` signal.
  - Any worker error terminates execution and notifies all loggers to record events, including `panic()` and recovery.
- Logger:
  - Provides `LoggerInterface` for custom loggers.
  - Predefined four log levels.
  - Default logger: Logs events to standard output and standard error.
  - Error collector: Gathers errors generated during workflow execution.

## Reference

For further details, please visit [this](https://docs.go-dag.dev.rho.social/basic-usage/use-simple.dag).
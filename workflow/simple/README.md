# Package simple

## Features

- Workflow:
  - No limit to the complexity of a directed acyclic graph.
  - Single input and single output.
  - Canceled at any time during execution.
  - Shipped with any loggers.
- Transit:
  - TransitInterface: Allow custom transits.
  - Any inputs and outputs(except the initial input and final output).
  - The worker accepts the incoming context parameter and allows receiving `Done()` signal.
  - Any worker error will terminate the execution and notify all the loggers to record events, including `panic()` and recover from it.
- Logger:
  - LoggerInterface: Allow custom loggers.
  - Predefined four log levels.
  - Default logger: Log events are output to the standard output and standard error.
  - Error collector: Collect errors generated during workflow execution.

## Reference

For more details, please visit [this](https://docs.go-dag.dev.rho.social/basic-usage/use-simple.dag).
# context

The Context needs to encompass three main functionalities:

- **Task Identification**: Each task execution requires a unique identifier to distinguish it from others.

- **Execution Parameters**: Parameters specific to the current execution, utilized by each transit, cache, and other components.

- **Execution Details**: Comprehensive information about the current execution, including the inputs and outputs of each transit,
error logging, resource consumption metrics (such as relative start time, duration, memory usage, and other resource consumption items), as well as the Directed Acyclic Graph (DAG) for the current execution.
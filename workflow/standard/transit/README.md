#  transit

The `transit` package focuses on managing workflows with fixed start and end nodes,
utilizing an Option pattern for graph initialization,
in contrast to the channel package's emphasis on channel-based data flow within directed acyclic graphs.

This package is designed to be highly flexible and interface-based,
allowing users to customize and tailor their workflow structures according to specific requirements.

## Background

The `transit` package introduces a `Graph` structure that differs from the `Graph` structure in the `channel` package in two key aspects:

1. **Option Pattern in NewGraph Method**: The `NewGraph` method in the `transit` package adopts the Option pattern, offering greater flexibility and configurability during graph initialization.

2. **Fixed Start and End Nodes**: In the `transit` package, the start and end nodes of the workflow, often referred to as the transit's entrance and exit points, are fixed as "input" and "output" respectively. Additional nodes can only be attached between these fixed start and end points using options.


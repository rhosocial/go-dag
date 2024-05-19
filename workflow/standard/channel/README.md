# channel

The `channel` package is designed to manage the channels between workflow transits. It provides a structure to handle directed acyclic graphs (DAGs) representing workflows, with nodes and edges indicating the flow and dependencies between different tasks or stages.

## Features

- **Graph Management**: Create and manage graphs representing workflows.
- **Cycle Detection**: Detect cycles within the graph to ensure the validity of workflows.
- **Node Management**: Add and manage nodes within the graph.
- **Customizable Options**: Easily configure the start and end nodes, as well as the list of nodes.

## Installation

To install the `channel` package, use:

```sh
go get github.com/rhosocial/go-dag/standard/channel

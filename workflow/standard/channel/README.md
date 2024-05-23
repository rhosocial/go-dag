# channel

The `channel` package is designed to manage the channels between workflow transits.
It provides a structure to handle directed acyclic graphs (DAGs) representing workflows,
with nodes and edges indicating the flow and dependencies between different tasks or stages.

The name "channel" for this package is chosen because it aligns with the central concept of utilizing channels
to facilitate the flow of data between different stages or tasks within a workflow.
Additionally, it serves to distinguish this package's functionality,
which focuses on managing the flow and dependencies represented by directed acyclic graphs (DAGs),
from other components such as the graph itself and the workflow transit.

This package has been fully interface-based. If the provided structures in the package do not meet your requirements,
you have the flexibility to implement your own structures based on the interfaces DAG, Node, and Transit.
On the other hand, you can also independently utilize this package for other purposes,
leveraging its functionality for managing directed acyclic graphs (DAGs) representing workflows or any other relevant scenarios.

## Features

- **Graph Management**: Create and manage graphs representing workflows.
- **Cycle Detection**: Detect cycles within the graph to ensure the validity of workflows.
- **Node Management**: Add and manage nodes within the graph.
- **Customizable Options**: Easily configure the start and end nodes, as well as the list of nodes.
- **Topological Sorting**: Perform topological sorting to determine a valid execution order of the nodes in the DAG.

## Installation

To install the `channel` package, use:

```sh
go get github.com/rhosocial/go-dag/workflow/standard/channel
```

## Usage

### Creating a Graph

You can create a graph by defining nodes and specifying their relationships.
Here's an example of how to create a simple graph:

```go
package main

import (
    "fmt"
    "github.com/rhosocial/go-dag/workflow/standard/channel"
)

func main() {
    nodes := []channel.Node{
        channel.NewNode("A", []string{}, []string{"B"}),
        channel.NewNode("B", []string{"A"}, []string{"C"}),
        channel.NewNode("C", []string{"B"}, []string{}),
    }

    graph, err := channel.NewGraph("A", "C", nodes...)
    if err != nil {
        fmt.Println("Error creating graph:", err)
        return
    }

    fmt.Printf("Graph created successfully: %v \n", graph)
}

```

### Cycle Detection

The `channel` package automatically checks for cycles when creating a graph.
If a cycle is detected, an error is returned:

```go
nodes := []channel.Node{
    channel.NewNode("A", []string{}, []string{"B"}),
    channel.NewNode("B", []string{"A"}, []string{"C"}),
    channel.NewNode("C", []string{"B"}, []string{"A"}), // Cycle here
}

_, err := channel.NewGraph("A", "C", nodes...)
if err != nil {
    fmt.Println("Error:", err)
}
```

### Topological Sorting

Topological sorting is used to determine a valid execution order of the nodes in the DAG.
This is crucial for workflows where certain tasks depend on the completion of others.

Here's how to perform a topological sort:

```go
sorts, err := graph.TopologicalSort()
if err != nil {
    fmt.Println("Error performing topological sort:", err)
    return
}

for i, sort := range sorts {
    fmt.Printf("Topological Sort %d: %v\n", i+1, sort)
}
```

The `TopologicalSort` method returns all possible topological sorts as a slice of slices of strings.
Each inner slice represents a valid execution order of the nodes.

For example:

```go
package main

import (
    "fmt"
    "github.com/rhosocial/go-dag/workflow/standard/channel"
)

func main() {
    nodes := []channel.Node{
        channel.NewNode("A", []string{}, []string{"B", "C"}),
        channel.NewNode("B", []string{"A"}, []string{"D"}),
        channel.NewNode("C", []string{"A"}, []string{"D"}),
        channel.NewNode("D", []string{"B", "C"}, []string{}),
    }

    graph, err := channel.NewGraph("A", "D", nodes...)
    if err != nil {
        fmt.Println("Error creating graph:", err)
        return
    }

    sorts, err := graph.TopologicalSort()
    if err != nil {
        fmt.Println("Error performing topological sort:", err)
        return
    }

    for i, sort := range sorts {
        fmt.Printf("Topological Sort %d: %v\n", i+1, sort)
    }
}
```

### Building a Graph from Transits

To build a graph from a list of `Transit` objects, use the `BuildGraphFromTransits` function.
You need to specify the source and sink node names:

```go
transits := []Transit{
    NewTransit("A", []string{}, []string{"chan1"}),
    NewTransit("B", []string{"chan1"}, []string{"chan2"}),
    NewTransit("C", []string{"chan2"}, []string{}),
}

graph, err := BuildGraphFromTransits("A", "C", transits...)
if err != nil {
    fmt.Printf("Error building graph: %v\n", err)
    return
}
```

For the generated graph, you can also check whether there are cycles and obtain all topological sorting.
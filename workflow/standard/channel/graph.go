// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package channel

// ReverseSlice reverses a slice of any type.
func ReverseSlice[T any](s []T) []T {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// DAG is an interface for graph operations.
type DAG interface {
	GetSourceName() string
	GetSinkName() string
	HasCycle() error
	TopologicalSort() ([][]string, error)
}

// Node is an interface representing a node in the graph.
type Node interface {
	GetName() string
	GetIncoming() []string
	AppendIncoming(...string)
	GetOutgoing() []string
	AppendOutgoing(...string)
}

// SimpleNode represents a node in the graph.
type SimpleNode struct {
	name     string
	incoming []string
	outgoing []string
}

// GetName returns the name of the node.
func (n *SimpleNode) GetName() string {
	return n.name
}

// GetIncoming returns the incoming edges of the node.
func (n *SimpleNode) GetIncoming() []string {
	return n.incoming
}

// AppendIncoming appends the incoming edge to the node.
func (n *SimpleNode) AppendIncoming(incoming ...string) {
	n.incoming = append(n.incoming, incoming...)
}

// GetOutgoing returns the outgoing edges of the node.
func (n *SimpleNode) GetOutgoing() []string {
	return n.outgoing
}

// AppendOutgoing appends the outgoing edge to the node.
func (n *SimpleNode) AppendOutgoing(outgoing ...string) {
	n.outgoing = append(n.outgoing, outgoing...)
}

// NewNode creates a new SimpleNode with the given name, incoming, and outgoing edges.
func NewNode(name string, incoming []string, outgoing []string) Node {
	return &SimpleNode{
		name:     name,
		incoming: incoming,
		outgoing: outgoing,
	}
}

// Graph represents the DAG
type Graph struct {
	NodesMap map[string]Node
	Source   string
	Sink     string
	DAG
}

// GetSourceName returns the source node name.
func (g *Graph) GetSourceName() string {
	return g.Source
}

// GetSinkName returns the sink node name.
func (g *Graph) GetSinkName() string {
	return g.Sink
}

// HasCycle checks if the graph contains a cycle and returns detailed cycle information.
func (g *Graph) HasCycle() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycleNodes []string

	var dfs func(nodeName string) bool
	dfs = func(nodeName string) bool {
		if recStack[nodeName] {
			cycleNodes = append(cycleNodes, nodeName)
			return true
		}
		if visited[nodeName] {
			return false
		}
		visited[nodeName] = true
		recStack[nodeName] = true

		for _, successor := range g.NodesMap[nodeName].GetOutgoing() {
			if dfs(successor) {
				cycleNodes = append(cycleNodes, nodeName)
				return true
			}
		}
		recStack[nodeName] = false
		return false
	}

	for nodeName := range g.NodesMap {
		if !visited[nodeName] {
			if dfs(nodeName) {
				return NewCycleError(ReverseSlice(cycleNodes)...)
			}
		}
	}

	return nil
}

// TopologicalSort returns all possible topological sorts of the DAG.
//
// The result is a two-dimensional slice, the length of which represents the number of possibilities.
// If a cycle exists in the graph, an error is reported.
//
// For example:
//
//	1.If the graph is A -> B -> C，[["A", "B", "C"]], nil will be returned.
//
//	2.If the graph is
//
//	   A -> B1 -> C
//
//	   |          |
//
//	    --> B2 -->
//
//	   [["A", "B1", "B2", "C"], ["A", "B2", "B1", "C"]], nil will be returned.
func (g *Graph) TopologicalSort() ([][]string, error) {
	// Check for cycle first
	if err := g.HasCycle(); err != nil {
		return nil, err
	}

	// Helper function to perform all topological sorts
	var allTopologicalSorts func(visited map[string]bool, inDegree map[string]int, stack []string) [][]string
	allTopologicalSorts = func(visited map[string]bool, inDegree map[string]int, stack []string) [][]string {
		var result [][]string
		flag := false

		for node, deg := range inDegree {
			if deg == 0 && !visited[node] {
				for _, successor := range g.NodesMap[node].GetOutgoing() {
					inDegree[successor]--
				}

				// Add node to stack and mark as visited
				stack = append(stack, node)
				visited[node] = true
				subResults := allTopologicalSorts(visited, inDegree, stack)

				// Collect results from recursive call
				for _, subResult := range subResults {
					result = append(result, subResult)
				}

				// Reset visited status and in-degree for backtracking
				visited[node] = false
				stack = stack[:len(stack)-1]
				for _, successor := range g.NodesMap[node].GetOutgoing() {
					inDegree[successor]++
				}

				flag = true
			}
		}

		// If flag is false, it means all nodes are processed
		if !flag {
			temp := make([]string, len(stack))
			copy(temp, stack)
			result = append(result, temp)
		}

		return result
	}

	// Initialize visited map and in-degree map
	visited := make(map[string]bool)
	inDegree := make(map[string]int)

	for node := range g.NodesMap {
		inDegree[node] = len(g.NodesMap[node].GetIncoming())
	}

	// Perform all topological sorts
	return allTopologicalSorts(visited, inDegree, []string{}), nil
}

// NewGraph initializes a new Graph.
//
// For the definition of a node, please see Node API.
// source and sink refers to the names of the starting and ending nodes of the graph.
func NewGraph(source, sink string, nodes ...Node) (*Graph, error) {
	graph := &Graph{
		NodesMap: make(map[string]Node),
		Source:   source,
		Sink:     sink,
	}

	// Add nodes to the graph
	for _, node := range nodes {
		graph.NodesMap[node.GetName()] = node
	}

	// Check for hanging predecessors and successors.
	for _, node := range nodes {
		if node.GetName() == source || node.GetName() == sink {
			continue
		}
		for _, pred := range node.GetIncoming() {
			if _, exists := graph.NodesMap[pred]; !exists {
				return nil, HangingIncomingError{predecessor: pred, name: node.GetName()}
			}
		}
		for _, successor := range node.GetOutgoing() {
			if _, exists := graph.NodesMap[successor]; !exists {
				return nil, HangingOutgoingError{successor: successor, name: node.GetName()}
			}
		}
	}

	var visitedSource Node
	var visitedSink Node

	for _, node := range nodes {
		if node.GetName() == source && len(node.GetIncoming()) == 0 {
			if visitedSource == nil {
				visitedSource = node
			} else {
				return nil, SourceDuplicatedError{source: visitedSource.GetName(), name: node.GetName()}
			}
		} else if node.GetName() == sink && len(node.GetOutgoing()) == 0 {
			if visitedSink == nil {
				visitedSink = node
			} else {
				return nil, SinkDuplicatedError{sink: visitedSink.GetName(), name: node.GetName()}
			}
		} else if len(node.GetIncoming()) == 0 {
			return nil, DanglingIncomingError{name: node.GetName()}
		} else if len(node.GetOutgoing()) == 0 {
			return nil, DanglingOutgoingError{name: node.GetName()}
		}
	}

	if err := graph.HasCycle(); err != nil {
		return nil, err
	}

	return graph, nil
}

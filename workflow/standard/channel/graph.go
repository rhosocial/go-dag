// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package channel

import (
	"fmt"
	"strings"
)

// CycleError represents an error due to a cycle in the graph.
type CycleError struct {
	nodes []string
	error
}

// Error returns a formatted string describing the cycle.
func (e *CycleError) Error() string {
	return fmt.Sprintf("graph has a cycle: %s", strings.Join(e.nodes, " -> "))
}

// NewCycleError creates a new CycleError with the given cycle nodes.
func NewCycleError(nodes ...string) *CycleError {
	return &CycleError{nodes: nodes}
}

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
}

// Node represents a node in the graph.
type Node struct {
	Name     string
	Incoming []string
	Outgoing []string
}

// NewNode creates a new Node with the given name, incoming, and outgoing edges.
func NewNode(name string, incoming []string, outgoing []string) *Node {
	return &Node{
		Name:     name,
		Incoming: incoming,
		Outgoing: outgoing,
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

		for _, successor := range g.NodesMap[nodeName].Outgoing {
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

// NewGraph initializes a new Graph.
func NewGraph(sourceName, sinkName string, nodes []*Node) (*Graph, error) {
	graph := &Graph{
		NodesMap: make(map[string]Node),
		Source:   sourceName,
		Sink:     sinkName,
	}

	// Add nodes to the graph
	for _, node := range nodes {
		graph.NodesMap[node.Name] = *node
	}

	// Check for hanging predecessors and successors.
	for _, node := range nodes {
		if node.Name == sourceName || node.Name == sinkName {
			continue
		}
		for _, pred := range node.Incoming {
			if _, exists := graph.NodesMap[pred]; !exists {
				return nil, HangingIncomingError{predecessor: pred, name: node.Name}
			}
		}
		for _, successor := range node.Outgoing {
			if _, exists := graph.NodesMap[successor]; !exists {
				return nil, HangingOutgoingError{successor: successor, name: node.Name}
			}
		}
	}

	var visitedSource *Node
	var visitedSink *Node

	for _, node := range nodes {
		if node.Name == sourceName && len(node.Incoming) == 0 {
			if visitedSource == nil {
				visitedSource = node
			} else {
				return nil, SourceDuplicatedError{source: visitedSource.Name, name: node.Name}
			}
		} else if node.Name == sinkName && len(node.Outgoing) == 0 {
			if visitedSink == nil {
				visitedSink = node
			} else {
				return nil, SinkDuplicatedError{sink: visitedSink.Name, name: node.Name}
			}
		} else if len(node.Incoming) == 0 {
			return nil, DanglingIncomingError{name: node.Name}
		} else if len(node.Outgoing) == 0 {
			return nil, DanglingOutgoingError{name: node.Name}
		}
	}

	if err := graph.HasCycle(); err != nil {
		return nil, err
	}

	return graph, nil
}

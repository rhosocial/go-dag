// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package channel

import (
	"fmt"
	"strings"
)

// ErrCycle represents an error due to a cycle in the graph.
type ErrCycle struct {
	cycleNodes []string
	error
}

// Error returns a formatted string describing the cycle.
func (e *ErrCycle) Error() string {
	return fmt.Sprintf("graph has a cycle: %s", strings.Join(e.cycleNodes, " -> "))
}

// NewErrCycle creates a new ErrCycle with the given cycle nodes.
func NewErrCycle(cycleNodes ...string) *ErrCycle {
	return &ErrCycle{cycleNodes: cycleNodes}
}

// reverse reverses a slice of any type.
func reverse[T any](s []T) []T {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// GraphInterface is an interface for graph operations.
type GraphInterface interface {
	GetSourceName() string
	GetSinkName() string
	HasCycle() error
}

// Node represents a node in the graph.
type Node struct {
	Name         string
	Predecessors []string
	Successors   []string
}

func NewNode(name string, predecessors []string, successors []string) *Node {
	return &Node{
		Name:         name,
		Predecessors: predecessors,
		Successors:   successors,
	}
}

// Graph represents the DAG
type Graph struct {
	Nodes      map[string]Node
	SourceName string
	SinkName   string
	GraphInterface
}

func (g *Graph) GetSourceName() string {
	return g.SourceName
}

func (g *Graph) GetSinkName() string {
	return g.SinkName
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

		for _, successor := range g.Nodes[nodeName].Successors {
			if dfs(successor) {
				cycleNodes = append(cycleNodes, nodeName)
				return true
			}
		}
		recStack[nodeName] = false
		return false
	}

	for nodeName := range g.Nodes {
		if !visited[nodeName] {
			if dfs(nodeName) {
				return NewErrCycle(reverse(cycleNodes)...)
			}
		}
	}

	return nil
}

type ErrHangingNodePredecessor struct {
	name        string
	predecessor string
	error
}

func (e ErrHangingNodePredecessor) Error() string {
	return fmt.Sprintf("node %s has a hanging predecessor %s", e.name, e.predecessor)
}

type ErrHangingNodeSuccessor struct {
	name      string
	successor string
	error
}

func (e ErrHangingNodeSuccessor) Error() string {
	return fmt.Sprintf("node %s has a hanging successor %s", e.name, e.successor)
}

type ErrDanglingNodePredecessor struct {
	name string
	error
}

func (e ErrDanglingNodePredecessor) Error() string {
	return fmt.Sprintf("node %s has no defined predecessor(s)", e.name)
}

type ErrDanglingNodeSuccessor struct {
	name string
	error
}

func (e ErrDanglingNodeSuccessor) Error() string {
	return fmt.Sprintf("node %s has no defined successor(s)", e.name)
}

type ErrSourceDuplicated struct {
	name   string
	source string
	error
}

func (e ErrSourceDuplicated) Error() string {
	return fmt.Sprintf("node %s cannot be source as the source %s already exists", e.name, e.source)
}

type ErrSinkDuplicated struct {
	name string
	sink string
	error
}

func (e ErrSinkDuplicated) Error() string {
	return fmt.Sprintf("node %s cannot be sink as the sink %s already exists", e.name, e.sink)
}

// NewGraph initializes a new Graph.
func NewGraph(sourceName, sinkName string, nodes []*Node) (*Graph, error) {
	graph := &Graph{
		Nodes:      make(map[string]Node),
		SourceName: sourceName,
		SinkName:   sinkName,
	}

	// Add nodes to the graph
	for _, node := range nodes {
		graph.Nodes[node.Name] = *node
	}

	// Check for hanging predecessors and successors.
	for _, node := range nodes {
		if node.Name == sourceName || node.Name == sinkName {
			continue
		}
		for _, pred := range node.Predecessors {
			if _, exists := graph.Nodes[pred]; !exists {
				return nil, ErrHangingNodePredecessor{predecessor: pred, name: node.Name}
			}
		}
		for _, successor := range node.Successors {
			if _, exists := graph.Nodes[successor]; !exists {
				return nil, ErrHangingNodeSuccessor{successor: successor, name: node.Name}
			}
		}
	}

	var visitedSource *Node
	var visitedSink *Node

	for _, node := range nodes {
		if node.Name == sourceName && len(node.Predecessors) == 0 {
			if visitedSource == nil {
				visitedSource = node
			} else {
				return nil, ErrSourceDuplicated{source: visitedSource.Name, name: node.Name}
			}
		} else if node.Name == sinkName && len(node.Successors) == 0 {
			if visitedSink == nil {
				visitedSink = node
			} else {
				return nil, ErrSinkDuplicated{sink: visitedSink.Name, name: node.Name}
			}
		} else if len(node.Predecessors) == 0 {
			return nil, ErrDanglingNodePredecessor{name: node.Name}
		} else if len(node.Successors) == 0 {
			return nil, ErrDanglingNodeSuccessor{name: node.Name}
		}
	}

	if err := graph.HasCycle(); err != nil {
		return nil, err
	}

	return graph, nil
}

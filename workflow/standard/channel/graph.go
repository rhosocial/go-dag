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
	GetStartNodeName() string
	GetEndNodeName() string
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
	Nodes         map[string]Node
	StartNodeName string
	EndNodeName   string
	GraphInterface
}

func (g *Graph) GetStartNodeName() string {
	return g.StartNodeName
}

func (g *Graph) GetEndNodeName() string {
	return g.EndNodeName
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

// NewGraph initializes a new Graph.
func NewGraph(startNodeName, endNodeName string, nodes []*Node) (*Graph, error) {
	graph := &Graph{
		Nodes:         make(map[string]Node),
		StartNodeName: startNodeName,
		EndNodeName:   endNodeName,
	}

	// Add nodes to the graph
	for _, node := range nodes {
		graph.Nodes[node.Name] = *node
	}

	// Check for hanging predecessors and successors.
	for _, node := range nodes {
		if node.Name == startNodeName || node.Name == endNodeName {
			continue
		}
		for _, pred := range node.Predecessors {
			if _, exists := graph.Nodes[pred]; !exists {
				return nil, fmt.Errorf("node %s has a hanging predecessor %s", node.Name, pred)
			}
		}
		for _, succeed := range node.Successors {
			if _, exists := graph.Nodes[succeed]; !exists {
				return nil, fmt.Errorf("node %s has a hanging successor %s", node.Name, succeed)
			}
		}
	}

	if err := graph.HasCycle(); err != nil {
		return nil, err
	}

	return graph, nil
}

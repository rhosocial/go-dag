// Copyright (c) 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package transit focuses on managing workflows with fixed start and end nodes,
// utilizing an Option pattern for graph initialization,
// in contrast to the channel package's emphasis on channel-based data flow within directed acyclic graphs.
package transit

import "github.com/rhosocial/go-dag/workflow/standard/channel"

const (
	InputNodeName  = "input"
	OutputNodeName = "output"
)

type GraphInterface interface {
	channel.DAG
	GetTransit() []channel.Transit
}

type Graph struct {
	graph    channel.DAG
	transits []channel.Transit
}

func (g *Graph) GetSourceName() string {
	return g.graph.GetSourceName()
}

func (g *Graph) GetSinkName() string {
	return g.graph.GetSinkName()
}

func (g *Graph) HasCycle() error {
	return g.graph.HasCycle()
}

func (g *Graph) TopologicalSort() ([][]string, error) {
	return g.graph.TopologicalSort()
}

func (g *Graph) GetTransit() []channel.Transit {
	return g.transits
}

// Option is a type for specifying options for creating a Transit.
type Option func(graph *Graph) error

// NewGraph creates a new Transit with the given input and output channels, and options.
func NewGraph(input, output string, options ...Option) (GraphInterface, error) {
	graph := &Graph{}

	// Apply options
	for _, option := range options {
		err := option(graph)
		if err != nil {
			return nil, err
		}
	}

	// Create input and output nodes
	inputTransit := channel.NewTransit(InputNodeName, []string{}, []string{input})
	outputTransit := channel.NewTransit(OutputNodeName, []string{output}, []string{})

	g, err := channel.BuildGraphFromTransits(InputNodeName, OutputNodeName, append(graph.transits, inputTransit, outputTransit)...)
	if err != nil {
		return nil, err
	}
	graph.graph = g

	return graph, nil
}

// WithIntermediateTransits is an option to add intermediate transits to the transit.
func WithIntermediateTransits(transits ...channel.Transit) Option {
	return func(g *Graph) error {
		if g.GetTransit() == nil {
			g.transits = make([]channel.Transit, 0)
		}
		g.transits = append(g.transits, transits...)
		return nil
	}
}

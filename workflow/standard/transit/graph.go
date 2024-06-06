// Copyright (c) 2023 - 2024. vistart. All rights reserved.
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

// GraphInterface defines the interface for managing the transit graph,
// incorporating operations from the channel package's DAG interface
// and an additional method to retrieve transits.
type GraphInterface interface {
	channel.DAG
	GetTransit() map[string]channel.Transit
}

// Graph represents the transit graph structure,
// composed of a channel.DAG instance and a slice of channel.Transit.
type Graph struct {
	channel.DAG
	transits map[string]channel.Transit
}

// GetTransit returns the transits attached to the transit graph.
func (g *Graph) GetTransit() map[string]channel.Transit {
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
	inputTransit := channel.NewTransit(InputNodeName, []string{}, []string{input}, nil)
	outputTransit := channel.NewTransit(OutputNodeName, []string{output}, []string{}, nil)

	// Build graph from transits
	s := make([]channel.Transit, 0, len(graph.transits))
	for _, t := range graph.transits {
		s = append(s, t)
	}
	g, err := channel.BuildGraphFromTransits(InputNodeName, OutputNodeName, append(s, inputTransit, outputTransit)...)
	if err != nil {
		return nil, err
	}
	graph.DAG = g

	return graph, nil
}

// WithIntermediateTransits is an option to add intermediate transits to the graph.
func WithIntermediateTransits(transits ...channel.Transit) Option {
	return func(g *Graph) error {
		if g.GetTransit() == nil {
			g.transits = make(map[string]channel.Transit)
		}
		for _, transit := range transits {
			g.transits[transit.GetName()] = transit
		}
		return nil
	}
}

// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package channel helps the workflow manage the transit channel.
package channel

// Interface defines the methods that need to be implemented for graph operations.
type Interface interface {
	ClearGraph()
	GetGraph() (DAG, error)
	GetChannelInput() string
	GetChannelOutput() string
	AppendNodes(...*Node)
}

// Channels contains the graph input, output, nodes, and the graph object.
type Channels struct {
	channelInput  string  // The name of the start node.
	channelOutput string  // The name of the end node.
	nodes         []*Node // A list of nodes.
	graph         DAG     // The implementation of the GraphInterface.
	Interface
}

// ClearGraph clears the current graph object.
func (c *Channels) ClearGraph() {
	c.graph = nil
}

// GetGraph returns the graph object, creating it if necessary.
func (c *Channels) GetGraph() (DAG, error) {
	if c.graph == nil {
		graph, err := NewGraph(c.channelInput, c.channelOutput, c.nodes)
		if err != nil {
			return nil, err
		}
		c.graph = graph
	}
	return c.graph, nil
}

func (c *Channels) GetChannelInput() string {
	return c.channelInput
}

func (c *Channels) GetChannelOutput() string {
	return c.channelOutput
}

func (c *Channels) AppendNodes(nodes ...*Node) {
	c.nodes = append(c.nodes, nodes...)
}

// Option defines a function type for configuring the Channels struct.
type Option func(*Channels) error

// NewChannels creates a new Channels instance and apply configuration options.
func NewChannels(options ...Option) (*Channels, error) {
	channels := &Channels{}
	for _, option := range options {
		err := option(channels)
		if err != nil {
			return nil, err
		}
	}
	return channels, nil
}

// WithChannelInput sets the start node name option.
func WithChannelInput(name string) Option {
	return func(channels *Channels) error {
		channels.channelInput = name
		return nil
	}
}

// WithChannelOutput sets the end node name option.
func WithChannelOutput(name string) Option {
	return func(channels *Channels) error {
		channels.channelOutput = name
		return nil
	}
}

const (
	DefaultChannelInput  = "input"  // Constant for the default start node name.
	DefaultChannelOutput = "output" // Constant for the default end node name.
)

// WithDefaultChannels sets the default start and end node names.
func WithDefaultChannels() Option {
	return func(channels *Channels) error {
		WithChannelInput(DefaultChannelInput)(channels)
		WithChannelOutput(DefaultChannelOutput)(channels)
		return nil
	}
}

// WithNodes sets the list of nodes option.
func WithNodes(nodes ...*Node) Option {
	return func(channels *Channels) error {
		channels.AppendNodes(nodes...)
		return nil
	}
}

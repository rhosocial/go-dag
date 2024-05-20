// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package channel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockChannels struct {
	Interface
}

func (m *MockChannels) GetGraph() (GraphInterface, error) {
	return nil, nil
}

func (m *MockChannels) ClearGraph() {}

func (m *MockChannels) GetChannelInput() string { return "" }

func (m *MockChannels) GetChannelOutput() string { return "" }

func (m *MockChannels) AppendNodes(...*Node) {}

var _ Interface = (*MockChannels)(nil)

func TestNewChannels(t *testing.T) {
	tests := []struct {
		name      string
		options   []Option
		expectErr bool
	}{
		{
			name: "Valid channels with default input/output",
			options: []Option{
				WithDefaultChannels(),
				WithNodes(
					&Node{Name: DefaultChannelInput, Successors: []string{"A"}},
					&Node{Name: "A", Predecessors: []string{DefaultChannelInput}, Successors: []string{"B"}},
					&Node{Name: "B", Predecessors: []string{"A"}, Successors: []string{DefaultChannelOutput}},
					&Node{Name: DefaultChannelOutput, Predecessors: []string{"B"}},
				),
			},
			expectErr: false,
		},
		{
			name: "Valid channels with custom input/output",
			options: []Option{
				WithChannelInput("start"),
				WithChannelOutput("end"),
				WithNodes(
					&Node{Name: "start", Successors: []string{"A"}},
					&Node{Name: "A", Predecessors: []string{"start"}, Successors: []string{"B"}},
					&Node{Name: "B", Predecessors: []string{"A"}, Successors: []string{"end"}},
					&Node{Name: "end", Predecessors: []string{"B"}},
				),
			},
			expectErr: false,
		},
		{
			name: "Invalid channels with cycle",
			options: []Option{
				WithDefaultChannels(),
				WithNodes(
					&Node{Name: DefaultChannelInput, Successors: []string{"A"}},
					&Node{Name: "A", Predecessors: []string{DefaultChannelInput}, Successors: []string{"B"}},
					&Node{Name: "B", Predecessors: []string{"A"}, Successors: []string{"C"}},
					&Node{Name: "C", Predecessors: []string{"B"}, Successors: []string{"A"}}, // Cycle here
					&Node{Name: DefaultChannelOutput, Predecessors: []string{"C"}},
				),
			},
			expectErr: true,
		},
		{
			name: "Invalid channels with hanging predecessor",
			options: []Option{
				WithDefaultChannels(),
				WithNodes(
					&Node{Name: DefaultChannelInput, Successors: []string{"A"}},
					&Node{Name: "A", Predecessors: []string{DefaultChannelInput}, Successors: []string{"B"}},
					&Node{Name: "B", Predecessors: []string{"A"}, Successors: []string{"output"}},
					&Node{Name: "C", Predecessors: []string{"D"}, Successors: []string{}}, // D does not exist
					&Node{Name: DefaultChannelOutput, Predecessors: []string{"C"}},
				),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channels, _ := NewChannels(tt.options...)
			_, err := channels.GetGraph()
			if (err != nil) != tt.expectErr {
				t.Errorf("NewChannels() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestChannels_GetGraph(t *testing.T) {
	channels, err := NewChannels(
		WithDefaultChannels(),
		WithNodes(
			&Node{Name: DefaultChannelInput, Successors: []string{"A"}},
			&Node{Name: "A", Predecessors: []string{DefaultChannelInput}, Successors: []string{"B"}},
			&Node{Name: "B", Predecessors: []string{"A"}, Successors: []string{DefaultChannelOutput}},
			&Node{Name: DefaultChannelOutput, Predecessors: []string{"B"}},
		),
	)
	if err != nil {
		t.Fatalf("NewChannels() error = %v", err)
	}

	graph, err := channels.GetGraph()
	if err != nil {
		t.Fatalf("Channels.GetGraph() error = %v", err)
	}

	if graph == nil {
		t.Fatal("Channels.GetGraph() returned nil graph")
	}

	err = graph.HasCycle()
	if err != nil {
		t.Errorf("Graph.HasCycle() error = %v", err)
	}
}

func TestChannels_ClearGraph(t *testing.T) {
	channels, err := NewChannels(
		WithDefaultChannels(),
		WithNodes(
			NewNode(DefaultChannelInput, nil, []string{"A"}),
			&Node{Name: "A", Predecessors: []string{DefaultChannelInput}, Successors: []string{"B"}},
			&Node{Name: "B", Predecessors: []string{"A"}, Successors: []string{DefaultChannelOutput}},
			&Node{Name: DefaultChannelOutput, Predecessors: []string{"B"}},
		),
	)
	if err != nil {
		t.Fatalf("NewChannels() error = %v", err)
	}

	_, err = channels.GetGraph()
	if err != nil {
		t.Fatalf("Channels.GetGraph() error = %v", err)
	}

	channels.ClearGraph()

	if channels.graph != nil {
		t.Fatal("Channels.ClearGraph() did not clear the graph")
	}
}

func WithOptionError() Option {
	return func(*Channels) error {
		return errors.New("failed to create channels")
	}
}

func TestChannelsWithOptionError(t *testing.T) {
	channels, err := NewChannels(WithOptionError())
	assert.Nil(t, channels)
	assert.Error(t, err)
}

func TestChannelsGetChannelInputOutput(t *testing.T) {
	channels, err := NewChannels(WithDefaultChannels())
	assert.NoError(t, err)
	assert.Equal(t, DefaultChannelInput, channels.GetChannelInput())
	assert.Equal(t, DefaultChannelOutput, channels.GetChannelOutput())
}

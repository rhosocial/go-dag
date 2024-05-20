// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package channel

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockGraph struct {
	GraphInterface
}

func (m MockGraph) GetSourceName() string {
	return ""
}

func (m MockGraph) GetSinkName() string {
	return ""
}

func (m MockGraph) HasCycle() error { return nil }

var _ GraphInterface = (*MockGraph)(nil)

func TestNewGraph(t *testing.T) {
	tests := []struct {
		name      string
		start     string
		end       string
		nodes     []*Node
		expectErr bool
	}{
		{
			name:  "Valid DAG",
			start: "input",
			end:   "output",
			nodes: []*Node{
				{Name: "input", Predecessors: []string{}, Successors: []string{"A"}},
				{Name: "A", Predecessors: []string{"input"}, Successors: []string{"B"}},
				{Name: "B", Predecessors: []string{"A"}, Successors: []string{"C"}},
				{Name: "C", Predecessors: []string{"B"}, Successors: []string{"output"}},
				{Name: "output", Predecessors: []string{"C"}, Successors: []string{}},
			},
			expectErr: false,
		},
		{
			name:  "Graph with a cycle",
			start: "input",
			end:   "output",
			nodes: []*Node{
				{Name: "input", Predecessors: []string{}, Successors: []string{"A"}},
				{Name: "A", Predecessors: []string{"input"}, Successors: []string{"B"}},
				{Name: "B", Predecessors: []string{"A"}, Successors: []string{"C"}},
				{Name: "C", Predecessors: []string{"B"}, Successors: []string{"A"}}, // Cycle here
				{Name: "output", Predecessors: []string{"C"}, Successors: []string{}},
			},
			expectErr: true,
		},
		{
			name:  "Graph with hanging predecessor",
			start: "input",
			end:   "output",
			nodes: []*Node{
				{Name: "input", Predecessors: []string{}, Successors: []string{"A"}},
				{Name: "A", Predecessors: []string{"input"}, Successors: []string{"B"}},
				{Name: "B", Predecessors: []string{"A"}, Successors: []string{"C"}},
				{Name: "C", Predecessors: []string{"B", "D"}, Successors: []string{"output"}}, // D does not exist
				{Name: "output", Predecessors: []string{"C"}, Successors: []string{}},
			},
			expectErr: true,
		},
		{
			name:  "Graph with hanging successor",
			start: "input",
			end:   "output",
			nodes: []*Node{
				{Name: "input", Predecessors: []string{}, Successors: []string{"A"}},
				{Name: "A", Predecessors: []string{"input"}, Successors: []string{"B"}},
				{Name: "B", Predecessors: []string{"A"}, Successors: []string{"C", "D"}}, // D does not exist
				{Name: "C", Predecessors: []string{"B"}, Successors: []string{"output"}},
				{Name: "output", Predecessors: []string{"C"}, Successors: []string{}},
			},
			expectErr: true,
		},
		{
			name:  "Single node input and output",
			start: "input",
			end:   "output",
			nodes: []*Node{
				{Name: "input", Predecessors: []string{}, Successors: []string{"output"}},
				{Name: "output", Predecessors: []string{"input"}, Successors: []string{}},
			},
			expectErr: false,
		},
		{
			name:  "Bad Case 1",
			start: "input",
			end:   "output",
			nodes: []*Node{
				{Name: "B", Predecessors: []string{"A"}},
				{Name: "C", Successors: []string{"output"}},
				{Name: "output", Predecessors: []string{"C"}},
				{Name: "input", Successors: []string{"A"}},
				{Name: "A", Predecessors: []string{"input"}, Successors: []string{"B"}},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := NewGraph(tt.start, tt.end, tt.nodes)
			if err == nil {
				assert.Equal(t, tt.start, graph.GetSourceName())
				assert.Equal(t, tt.end, graph.GetSinkName())
			}
			if (err != nil) != tt.expectErr {
				t.Errorf("NewGraph() error = %v, expectErr %v", err, tt.expectErr)
			} else if tt.expectErr && err != nil {
				fmt.Println(err)
			}
		})
	}
}

func TestHasCycle(t *testing.T) {
	tests := []struct {
		name      string
		graph     *Graph
		expectErr bool
	}{
		{
			name: "No cycle",
			graph: &Graph{
				Nodes: map[string]Node{
					"input":  {Name: "input", Successors: []string{"A"}},
					"A":      {Name: "A", Predecessors: []string{"input"}, Successors: []string{"B"}},
					"B":      {Name: "B", Predecessors: []string{"A"}, Successors: []string{"C"}},
					"C":      {Name: "C", Predecessors: []string{"B"}, Successors: []string{"output"}},
					"output": {Name: "output", Predecessors: []string{"C"}},
				},
				SourceName: "input",
				SinkName:   "output",
			},
			expectErr: false,
		},
		{
			name: "Cycle present",
			graph: &Graph{
				Nodes: map[string]Node{
					"input":  {Name: "input", Successors: []string{"A"}},
					"A":      {Name: "A", Predecessors: []string{"input"}, Successors: []string{"B"}},
					"B":      {Name: "B", Predecessors: []string{"A"}, Successors: []string{"C"}},
					"C":      {Name: "C", Predecessors: []string{"B"}, Successors: []string{"A"}},
					"output": {Name: "output", Predecessors: []string{"C"}},
				},
				SourceName: "input",
				SinkName:   "output",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.graph.HasCycle()
			if (err != nil) != tt.expectErr {
				t.Errorf("HasCycle() error = %v, expectErr %v", err, tt.expectErr)
			} else if tt.expectErr && err != nil {
				fmt.Println(err)
			}
		})
	}
}

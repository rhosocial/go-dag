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
	DAG
}

func (m MockGraph) GetSourceName() string {
	return ""
}

func (m MockGraph) GetSinkName() string {
	return ""
}

func (m MockGraph) HasCycle() error { return nil }

func (m MockGraph) TopologicalSort() ([][]string, error) { return nil, nil }

var _ DAG = (*MockGraph)(nil)

func TestNewGraph(t *testing.T) {
	tests := []struct {
		name      string
		start     string
		end       string
		nodes     []Node
		expectErr bool
	}{
		{
			name:  "Valid DAG",
			start: "input",
			end:   "output",
			nodes: []Node{
				NewNode("input", []string{}, []string{"A"}),
				NewNode("A", []string{"input"}, []string{"B"}),
				NewNode("B", []string{"A"}, []string{"C"}),
				NewNode("C", []string{"B"}, []string{"output"}),
				NewNode("output", []string{"C"}, []string{}),
			},
			expectErr: false,
		},
		{
			name:  "Graph with a cycle",
			start: "input",
			end:   "output",
			nodes: []Node{
				NewNode("input", []string{}, []string{"A"}),
				NewNode("A", []string{"input"}, []string{"B"}),
				NewNode("B", []string{"A"}, []string{"C"}),
				NewNode("C", []string{"B"}, []string{"A"}), // Cycle here
				NewNode("output", []string{"C"}, []string{}),
			},
			expectErr: true,
		},
		{
			name:  "Graph with hanging predecessor",
			start: "input",
			end:   "output",
			nodes: []Node{
				NewNode("input", []string{}, []string{"A"}),
				NewNode("A", []string{"input"}, []string{"B"}),
				NewNode("B", []string{"A"}, []string{"C"}),
				NewNode("C", []string{"B", "D"}, []string{"output"}), // D does not exist
				NewNode("output", []string{"C"}, []string{}),
			},
			expectErr: true,
		},
		{
			name:  "Graph with hanging successor",
			start: "input",
			end:   "output",
			nodes: []Node{
				NewNode("input", []string{}, []string{"A"}),
				NewNode("A", []string{"input"}, []string{"B"}),
				NewNode("B", []string{"A"}, []string{"C", "D"}), // D does not exist
				NewNode("C", []string{"B"}, []string{"output"}),
				NewNode("output", []string{"C"}, []string{}),
			},
			expectErr: true,
		},
		{
			name:  "Single node input and output",
			start: "input",
			end:   "output",
			nodes: []Node{
				NewNode("input", []string{}, []string{"output"}),
				NewNode("output", []string{"input"}, []string{}),
			},
			expectErr: false,
		},
		{
			name:  "Bad Case 1",
			start: "input",
			end:   "output",
			nodes: []Node{
				NewNode("B", []string{"A"}, []string{}),
				NewNode("C", []string{}, []string{"output"}),
				NewNode("output", []string{"C"}, []string{}),
				NewNode("input", []string{}, []string{"A"}),
				NewNode("A", []string{"input"}, []string{"B"}),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := NewGraph(tt.start, tt.end, tt.nodes...)
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
				NodesMap: map[string]Node{
					"input":  NewNode("input", []string{}, []string{"A"}),
					"A":      NewNode("A", []string{"input"}, []string{"B"}),
					"B":      NewNode("B", []string{"A"}, []string{"C"}),
					"C":      NewNode("C", []string{"B"}, []string{"output"}),
					"output": NewNode("output", []string{"C"}, []string{}),
				},
				Source: "input",
				Sink:   "output",
			},
			expectErr: false,
		},
		{
			name: "Cycle present",
			graph: &Graph{
				NodesMap: map[string]Node{
					"input":  NewNode("input", []string{}, []string{"A"}),
					"A":      NewNode("A", []string{"input"}, []string{"B"}),
					"B":      NewNode("B", []string{"A"}, []string{"C"}),
					"C":      NewNode("C", []string{"B"}, []string{"A"}),
					"output": NewNode("output", []string{"C"}, []string{}),
				},
				Source: "input",
				Sink:   "output",
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

func TestNewGraphWithExamples(t *testing.T) {
	t.Run("Example 1: Simple DAG without cycles or dangling nodes", func(t *testing.T) {
		// Example 1: Simple DAG without cycles or dangling nodes
		nodes1 := []Node{
			NewNode("input", []string{}, []string{"A"}),
			NewNode("A", []string{"input"}, []string{"B"}),
			NewNode("B", []string{"A"}, []string{"output"}),
			NewNode("output", []string{"B"}, []string{}),
		}

		graph1, err1 := NewGraph("input", "output", nodes1...)
		if err1 != nil {
			t.Fatalf("Error creating graph from nodes1: %v", err1)
		}
		t.Log("Graph created successfully from nodes1:", graph1)
	})

	t.Run("Example 2: Graph with a cycle", func(t *testing.T) {
		// Example 2: Graph with a cycle
		nodes2 := []Node{
			NewNode("input", []string{}, []string{"A"}),
			NewNode("A", []string{"input"}, []string{"B"}),
			NewNode("B", []string{"A"}, []string{"A"}), // Creates a cycle
			NewNode("output", []string{"B"}, []string{}),
		}

		_, err2 := NewGraph("input", "output", nodes2...)
		if err2 == nil {
			t.Fatal("Expected error creating graph from nodes2 due to cycle, but got none")
		}
		t.Log("Error creating graph from nodes2 (expected cycle):", err2)
	})

	t.Run("Example 3: Graph with a dangling outgoing node", func(t *testing.T) {
		// Example 3: Graph with a dangling outgoing node
		nodes3 := []Node{
			NewNode("input", []string{}, []string{"A"}),
			NewNode("A", []string{"input"}, []string{"B"}),
			NewNode("B", []string{"A"}, []string{}),      // B is a dangling node
			NewNode("output", []string{"C"}, []string{}), // C is not defined
		}

		_, err3 := NewGraph("input", "output", nodes3...)
		if err3 == nil {
			t.Fatal("Expected error creating graph from nodes3 due to dangling outgoing node, but got none")
		}
		t.Log("Error creating graph from nodes3 (expected dangling outgoing node):", err3)
	})

	t.Run("Example 4: Graph with a dangling incoming node", func(t *testing.T) {
		// Example 4: Graph with a dangling incoming node
		nodes4 := []Node{
			NewNode("input", []string{}, []string{"A"}),
			NewNode("A", []string{"input"}, []string{"B"}),
			NewNode("B", []string{}, []string{"A"}),      // B is a dangling node
			NewNode("output", []string{"C"}, []string{}), // C is not defined
		}

		_, err4 := NewGraph("input", "output", nodes4...)
		if err4 == nil {
			t.Fatal("Expected error creating graph from nodes4 due to dangling incoming node, but got none")
		}
		t.Log("Error creating graph from nodes4 (expected dangling outgoing node):", err4)
	})

	t.Run("Example 5: Graph with both cycle and dangling nodes", func(t *testing.T) {
		// Example 5: Graph with both cycle and dangling nodes
		nodes5 := []Node{
			NewNode("input", []string{}, []string{"A"}),
			NewNode("A", []string{"input"}, []string{"B"}),
			NewNode("B", []string{"A"}, []string{"input"}), // Creates a cycle
			NewNode("output", []string{"D"}, []string{}),   // D is not defined
		}

		_, err5 := NewGraph("input", "output", nodes5...)
		if err5 == nil {
			t.Fatal("Expected error creating graph from nodes5 due to cycle and dangling node, but got none")
		}
		t.Log("Error creating graph from nodes5 (expected cycle and dangling node):", err5)
	})

	t.Run("Example 6: Graph with duplicated sources", func(t *testing.T) {
		// Example 6: Graph with duplicated sources
		nodes6 := []Node{
			NewNode("input", []string{}, []string{"A"}),
			NewNode("A", []string{"input"}, []string{"B"}),
			NewNode("B", []string{"A"}, []string{"output"}),
			NewNode("input", []string{}, []string{"C"}), // Duplicated source
			NewNode("output", []string{"B"}, []string{}),
		}

		_, err6 := NewGraph("input", "output", nodes6...)
		if err6 == nil {
			t.Fatal("Expected error creating graph from nodes6 due to duplicated source, but got none")
		}
		t.Log("Error creating graph from nodes6 (expected duplicated source):", err6)
	})

	t.Run("Example 7: Graph with duplicated sinks", func(t *testing.T) {
		// Example 7: Graph with duplicated sinks
		nodes7 := []Node{
			NewNode("input", []string{}, []string{"A"}),
			NewNode("A", []string{"input"}, []string{"B"}),
			NewNode("B", []string{"A"}, []string{"output"}),
			NewNode("output", []string{"B"}, []string{}),
			NewNode("output", []string{"C"}, []string{}), // Duplicated sink
		}

		_, err7 := NewGraph("input", "output", nodes7...)
		if err7 == nil {
			t.Fatal("Expected error creating graph from nodes7 due to duplicated sink, but got none")
		}
		t.Log("Error creating graph from nodes7 (expected duplicated sink):", err7)
	})

	t.Run("Example 8: Valid graph with input and output nodes", func(t *testing.T) {
		// Example 8: Valid graph with input and output nodes
		nodes8 := []Node{
			NewNode("input", []string{}, []string{"A"}),
			NewNode("A", []string{"input"}, []string{"B"}),
			NewNode("B", []string{"A"}, []string{"C"}),
			NewNode("C", []string{"B"}, []string{"output"}),
			NewNode("output", []string{"C"}, []string{}),
		}

		graph8, err8 := NewGraph("input", "output", nodes8...)
		if err8 != nil {
			t.Fatalf("Error creating graph from nodes8: %v", err8)
		}
		t.Log("Graph created successfully from nodes8:", graph8)
	})
}

func TestGraph_TopologicalSort(t *testing.T) {
	t.Run("Simple DAG", func(t *testing.T) {
		nodes := []Node{
			NewNode("A", []string{}, []string{"B"}),
			NewNode("B", []string{"A"}, []string{"C"}),
			NewNode("C", []string{"B"}, []string{}),
		}

		graph, err := NewGraph("A", "C", nodes...)
		assert.NoError(t, err)
		assert.NotNil(t, graph)

		sorts, err := graph.TopologicalSort()
		assert.NoError(t, err)
		assert.Equal(t, [][]string{
			{"A", "B", "C"},
		}, sorts)
	})

	t.Run("DAG with Multiple Branches", func(t *testing.T) {
		nodes := []Node{
			NewNode("A", []string{}, []string{"B", "C"}),
			NewNode("B", []string{"A"}, []string{"D"}),
			NewNode("C", []string{"A"}, []string{"D"}),
			NewNode("D", []string{"B", "C"}, []string{}),
		}

		graph, err := NewGraph("A", "D", nodes...)
		assert.NoError(t, err)
		assert.NotNil(t, graph)

		sorts, err := graph.TopologicalSort()
		assert.NoError(t, err)
		assert.ElementsMatch(t, [][]string{
			{"A", "B", "C", "D"},
			{"A", "C", "B", "D"},
		}, sorts)
	})

	t.Run("Single Node DAG", func(t *testing.T) {
		nodes := []Node{
			NewNode("A", []string{}, []string{}),
		}

		graph, err := NewGraph("A", "A", nodes...)
		assert.NoError(t, err)
		assert.NotNil(t, graph)

		sorts, err := graph.TopologicalSort()
		assert.NoError(t, err)
		assert.Equal(t, [][]string{
			{"A"},
		}, sorts)
	})

	t.Run("DAG with Cycle", func(t *testing.T) {
		nodes := []Node{
			NewNode("A", []string{}, []string{"B"}),
			NewNode("B", []string{"A"}, []string{"C"}),
			NewNode("C", []string{"B"}, []string{"A"}), // Cycle here
		}

		graph := &Graph{
			NodesMap: make(map[string]Node),
			Source:   "A",
			Sink:     "C",
		}

		// Add nodes to the graph
		for _, node := range nodes {
			graph.NodesMap[node.GetName()] = node
		}
		_, err := graph.TopologicalSort()
		assert.Error(t, err)
	})

	t.Run("Empty Graph", func(t *testing.T) {
		var nodes []Node

		graph, err := NewGraph("", "", nodes...)
		assert.NoError(t, err)
		assert.NotNil(t, graph)

		sorts, err := graph.TopologicalSort()
		assert.NoError(t, err)
		assert.Equal(t, [][]string{{}}, sorts) // Expecting an empty slice of slices
	})

	t.Run("Nil Graph", func(t *testing.T) {
		graph, err := NewGraph("", "")
		assert.NoError(t, err)
		assert.NotNil(t, graph)

		sorts, err := graph.TopologicalSort()
		assert.NoError(t, err)
		assert.Equal(t, [][]string{{}}, sorts) // Expecting an empty slice of slices
	})
}
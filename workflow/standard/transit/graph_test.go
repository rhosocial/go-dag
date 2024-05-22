// Copyright (c) 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package transit

import (
	"errors"
	"testing"

	"github.com/rhosocial/go-dag/workflow/standard/channel"
	"github.com/stretchr/testify/assert"
)

func WithTransitsError() Option {
	return func(_ *Graph) error {
		return errors.New("error")
	}
}

func TestNewGraph(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		graph, err := NewGraph("input", "output", WithIntermediateTransits(
			channel.NewTransit("t1", []string{"input"}, []string{"output"}),
		))
		assert.NoError(t, err)
		t.Log(graph)
		sorts, err := graph.TopologicalSort()
		t.Log(sorts, err)
		err = graph.HasCycle()
		assert.NoError(t, err)
	})
	t.Run("with option error", func(t *testing.T) {
		graph, err := NewGraph("input", "output", WithTransitsError())
		assert.Error(t, err)
		assert.Nil(t, graph)
	})
}

// Helper function to create and validate graph from transits
func createAndValidateGraph(t *testing.T, transits []channel.Transit, sourceName, sinkName string) GraphInterface {
	graph, err := NewGraph(sourceName, sinkName, WithIntermediateTransits(transits...))
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}
	if err := graph.HasCycle(); err != nil {
		t.Fatalf("graph has a cycle: %v", err)
	}
	return graph
}

func TestLinearDAG(t *testing.T) {
	transits := []channel.Transit{
		channel.NewTransit("A", []string{"input"}, []string{"chan1"}),
		channel.NewTransit("B", []string{"chan1"}, []string{"chan2"}),
		channel.NewTransit("C", []string{"chan2"}, []string{"output"}),
	}
	graph := createAndValidateGraph(t, transits, "input", "output")
	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("topological sort failed: %v", err)
	}
	expected := [][]string{
		{"input", "A", "B", "C", "output"},
	}
	validateTopologicalSort(t, sorted, expected)
}

func TestComplexDAG(t *testing.T) {
	transits := []channel.Transit{
		channel.NewTransit("A", []string{"input"}, []string{"chan1"}),
		channel.NewTransit("B", []string{"chan1"}, []string{"chan2", "chan3"}),
		channel.NewTransit("C", []string{"chan2"}, []string{"chan4"}),
		channel.NewTransit("D", []string{"chan3"}, []string{"chan4"}),
		channel.NewTransit("E", []string{"chan4"}, []string{"output"}),
	}
	graph := createAndValidateGraph(t, transits, "input", "output")
	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("topological sort failed: %v", err)
	}
	expected := [][]string{
		{"input", "A", "B", "C", "D", "E", "output"},
		{"input", "A", "B", "D", "C", "E", "output"},
	}
	validateTopologicalSort(t, sorted, expected)
}

func TestCycleDetection(t *testing.T) {
	transits := []channel.Transit{
		channel.NewTransit("A", []string{"input"}, []string{"chan1"}),
		channel.NewTransit("B", []string{"chan1"}, []string{"chan2"}),
		channel.NewTransit("C", []string{"chan2"}, []string{"chan3"}),
		channel.NewTransit("D", []string{"chan3"}, []string{"chan1"}),
	}
	_, err := NewGraph("input", "output", WithIntermediateTransits(transits...))
	if err == nil {
		t.Fatal("expected cycle detection error, but got none")
	}
	var cycleError *channel.CycleError
	if !errors.As(err, &cycleError) {
		t.Fatalf("expected CycleError, but got %v", err)
	}
}

func TestDanglingIncoming(t *testing.T) {
	transits := []channel.Transit{
		channel.NewTransit("A", []string{"input"}, []string{"chan1"}),
		channel.NewTransit("B", []string{"chan1"}, []string{}),
		channel.NewTransit("C", []string{"chan2"}, []string{}), // Dangling incoming chan2
	}
	_, err := NewGraph("input", "output", WithIntermediateTransits(transits...))
	if err == nil {
		t.Fatal("expected dangling incoming error, but got none")
	}
}

func TestDanglingOutgoing(t *testing.T) {
	transits := []channel.Transit{
		channel.NewTransit("A", []string{}, []string{"chan1"}),
		channel.NewTransit("B", []string{"chan1"}, []string{}),
		channel.NewTransit("C", []string{}, []string{"chan2"}), // Dangling outgoing chan2
	}
	_, err := NewGraph("input", "output", WithIntermediateTransits(transits...))
	if err == nil {
		t.Fatal("expected dangling outgoing error, but got none")
	}
}

// Helper function to validate topological sort results
func validateTopologicalSort(t *testing.T, actual, expected [][]string) {
	if len(actual) != len(expected) {
		t.Fatalf("expected %d topological sorts, but got %d", len(expected), len(actual))
	}
	for _, exp := range expected {
		found := false
		for _, act := range actual {
			if equalSlices(act, exp) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected topological sort %v not found in actual results", exp)
		}
	}
}

// Utility function to check if two slices are equal
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGraph_GetSourceName(t *testing.T) {
	transits := []channel.Transit{
		channel.NewTransit("A", []string{"input"}, []string{"chan1"}),
		channel.NewTransit("B", []string{"chan1"}, []string{"chan2"}),
		channel.NewTransit("C", []string{"chan2"}, []string{"output"}),
	}
	graph := createAndValidateGraph(t, transits, "input", "output")
	assert.Equal(t, "input", graph.GetSourceName())
}

func TestGraph_GetSinkName(t *testing.T) {
	transits := []channel.Transit{
		channel.NewTransit("A", []string{"input"}, []string{"chan1"}),
		channel.NewTransit("B", []string{"chan1"}, []string{"chan2"}),
		channel.NewTransit("C", []string{"chan2"}, []string{"output"}),
	}
	graph := createAndValidateGraph(t, transits, "input", "output")
	assert.Equal(t, "output", graph.GetSinkName())
}

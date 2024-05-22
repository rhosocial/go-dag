// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package channel

import (
	"errors"
	"testing"
)

// Helper function to create and validate graph from transits
func createAndValidateGraph(t *testing.T, transits []Transit, sourceName, sinkName string) *Graph {
	graph, err := BuildGraphFromTransits(sourceName, sinkName, transits...)
	if err != nil {
		t.Fatalf("failed to build graph: %v", err)
	}
	if err := graph.HasCycle(); err != nil {
		t.Fatalf("graph has a cycle: %v", err)
	}
	return graph
}

func TestLinearDAG(t *testing.T) {
	transits := []Transit{
		NewTransit("A", []string{}, []string{"chan1"}),
		NewTransit("B", []string{"chan1"}, []string{"chan2"}),
		NewTransit("C", []string{"chan2"}, []string{}),
	}
	graph := createAndValidateGraph(t, transits, "A", "C")
	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("topological sort failed: %v", err)
	}
	expected := [][]string{
		{"A", "B", "C"},
	}
	validateTopologicalSort(t, sorted, expected)
}

func TestComplexDAG(t *testing.T) {
	transits := []Transit{
		NewTransit("A", []string{}, []string{"chan1"}),
		NewTransit("B", []string{"chan1"}, []string{"chan2", "chan3"}),
		NewTransit("C", []string{"chan2"}, []string{"chan4"}),
		NewTransit("D", []string{"chan3"}, []string{"chan4"}),
		NewTransit("E", []string{"chan4"}, []string{}),
	}
	graph := createAndValidateGraph(t, transits, "A", "E")
	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("topological sort failed: %v", err)
	}
	expected := [][]string{
		{"A", "B", "C", "D", "E"},
		{"A", "B", "D", "C", "E"},
	}
	validateTopologicalSort(t, sorted, expected)
}

func TestCycleDetection(t *testing.T) {
	transits := []Transit{
		NewTransit("A", []string{}, []string{"chan1"}),
		NewTransit("B", []string{"chan1"}, []string{"chan2"}),
		NewTransit("C", []string{"chan2"}, []string{"chan3"}),
		NewTransit("D", []string{"chan3"}, []string{"chan1"}),
	}
	_, err := BuildGraphFromTransits("A", "D", transits...)
	if err == nil {
		t.Fatal("expected cycle detection error, but got none")
	}
	var cycleError *CycleError
	if !errors.As(err, &cycleError) {
		t.Fatalf("expected CycleError, but got %v", err)
	}
}

func TestDanglingIncoming(t *testing.T) {
	transits := []Transit{
		NewTransit("A", []string{}, []string{"chan1"}),
		NewTransit("B", []string{"chan1"}, []string{}),
		NewTransit("C", []string{"chan2"}, []string{}), // Dangling incoming chan2
	}
	_, err := BuildGraphFromTransits("A", "B", transits...)
	if err == nil {
		t.Fatal("expected dangling incoming error, but got none")
	}
}

func TestDanglingOutgoing(t *testing.T) {
	transits := []Transit{
		NewTransit("A", []string{}, []string{"chan1"}),
		NewTransit("B", []string{"chan1"}, []string{}),
		NewTransit("C", []string{}, []string{"chan2"}), // Dangling outgoing chan2
	}
	_, err := BuildGraphFromTransits("A", "B", transits...)
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

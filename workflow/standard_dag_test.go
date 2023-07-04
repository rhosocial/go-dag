package workflow

import "testing"

func TestNewStandardDAG(t *testing.T) {
	dag := NewStandardDAG[int, int]()
	t.Log(dag)
}

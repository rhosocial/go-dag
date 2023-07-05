package workflow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStandardDAG(t *testing.T) {
	root := context.Background()

	dag := NewStandardDAG[int, int]()
	assert.NotNil(t, dag)

	input := 1
	output := dag.RunOnce(root, &input)
	assert.Nil(t, output)
}

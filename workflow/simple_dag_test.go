package workflow

import (
	"errors"
	"log"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DAGTwoParallelTransits struct {
	SimpleDAG[string, string]
}

// NewDAGTwoParallelTransits defines a workflow.
func NewDAGTwoParallelTransits() *DAGTwoParallelTransits {
	f := DAGTwoParallelTransits{
		SimpleDAG: *NewSimpleDAG[string, string](),
	}
	f.InitChannels("input", "t11", "t12", "t21", "t22", "output")
	//         input               t11              output
	// input ----+----> transit1 -------> transit ----------> output
	//           |                 t12       ^
	//           +----> transit2 ------------+
	f.InitWorkflow("input", "output", &SimpleDAGWorkflowTransit{
		name:           "input",
		channelInputs:  []string{"input"},
		channelOutputs: []string{"t11", "t12"},
		worker: func(a ...any) (any, error) {
			return a[0], nil
		},
	}, &SimpleDAGWorkflowTransit{
		name:           "transit1",
		channelInputs:  []string{"t11"},
		channelOutputs: []string{"t21"},
		worker: func(a ...any) (any, error) {
			return a[0], nil
		},
	}, &SimpleDAGWorkflowTransit{
		name:           "transit2",
		channelInputs:  []string{"t12"},
		channelOutputs: []string{"t22"},
		worker: func(a ...any) (any, error) {
			var e = SimpleDAGValueTypeError{actual: new(int), expect: new(string)}
			var t = SimpleDAGValueTypeError{actual: new(int), expect: new(string)}
			log.Println(errors.Is(&e, &t))
			return a[0], nil
		},
	}, &SimpleDAGWorkflowTransit{
		name:           "transit",
		channelInputs:  []string{"t21", "t22"},
		channelOutputs: []string{"output"},
		worker: func(a ...any) (any, error) {
			var r string
			for i, c := range a {
				r = r + strconv.Itoa(i) + c.(string)
			}
			return r, nil
		},
	})
	return &f
}

func TestDAGTwoParallelTransits(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	t.Run("run successfully", func(t *testing.T) {
		f := NewDAGTwoParallelTransits()
		var input = "test"
		for i := 0; i < 5; i++ {
			var results = f.Execute(&input)
			assert.Equal(t, "0test1test", *results)
		}
		var results = f.RunOnce(&input)
		assert.Equal(t, "0test1test", *results)
	})

	t.Run("closed channel and run again", func(t *testing.T) {
		defer func() { recover() }()
		f := NewDAGTwoParallelTransits()
		var input = "test"
		var results = f.RunOnce(&input)
		assert.Equal(t, "0test1test", *results)
		f.RunOnce(&input)
	})
}

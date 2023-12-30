package workflow

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type SimpleDAG1 struct {
	SimpleDAG[string, string]
}

func TestSimpleDAGChannel_Exists(t *testing.T) {
	f := SimpleDAG1{
		SimpleDAG: *NewSimpleDAG[string, string](),
	}
	assert.False(t, f.Exists("input"))
	assert.False(t, f.Exists("output"))
	f.InitChannels("input", "output")
	assert.True(t, f.Exists("input"))
	assert.True(t, f.Exists("output"))
}

func TestSimpleDAGValueTypeError_Error(t *testing.T) {
	f := SimpleDAG1{
		SimpleDAG: *NewSimpleDAG[string, string](),
	}
	f.InitChannels("input", "t11", "output")
	f.InitWorkflow("input", "output", &SimpleDAGWorkflowTransit{
		name:           "input",
		channelInputs:  []string{"input"},
		channelOutputs: []string{"t11"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			return a[0], nil
		},
	}, &SimpleDAGWorkflowTransit{
		name:           "output",
		channelInputs:  []string{"t11"},
		channelOutputs: []string{"output"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			return 0, nil
		},
	})
	input := "input"
	assert.Nil(t, f.Execute(context.Background(), &input))
	assert.Nil(t, f.RunOnce(context.Background(), &input))
}

// NewDAGTwoParallelTransits defines a workflow.
func NewDAGTwoParallelTransits() *SimpleDAG1 {
	f := SimpleDAG1{
		SimpleDAG: *NewSimpleDAG[string, string](),
	}
	f.InitChannels("input", "t11", "t12", "t21", "t22", "output")
	//         t:input          t:transit1          t:output
	//c:input ----+----> c:t11 ------------> c:t21 -----+----> c:output
	//            |             t:transit2              ^
	//            +----> c:t12 ------------> c:t22 -----+
	f.InitWorkflow("input", "output", &SimpleDAGWorkflowTransit{
		name:           "input",
		channelInputs:  []string{"input"},
		channelOutputs: []string{"t11", "t12"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			return a[0], nil
		},
	}, &SimpleDAGWorkflowTransit{
		name:           "transit1",
		channelInputs:  []string{"t11"},
		channelOutputs: []string{"t21"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			return a[0], nil
		},
	}, &SimpleDAGWorkflowTransit{
		name:           "transit2",
		channelInputs:  []string{"t12"},
		channelOutputs: []string{"t22"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			//var e = SimpleDAGValueTypeError{actual: new(int), expect: new(string)}
			//var t = SimpleDAGValueTypeError{actual: new(int), expect: new(string)}
			//log.Println(errors.Is(&e, &t))
			return a[0], nil
		},
	}, &SimpleDAGWorkflowTransit{
		name:           "transit",
		channelInputs:  []string{"t21", "t22"},
		channelOutputs: []string{"output"},
		worker: func(ctx context.Context, a ...any) (any, error) {
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
	root := context.Background()
	t.Run("run successfully", func(t *testing.T) {
		f := NewDAGTwoParallelTransits()
		var input = "test"
		for i := 0; i < 5; i++ {
			var results = f.Execute(root, &input)
			assert.Equal(t, "0test1test", *results)
		}
		var results = f.RunOnce(root, &input)
		assert.Equal(t, "0test1test", *results)
	})

	t.Run("closed channel and run again", func(t *testing.T) {
		defer func() { recover() }()
		f := NewDAGTwoParallelTransits()
		var input = "test"
		var results = f.RunOnce(root, &input)
		assert.Equal(t, "0test1test", *results)
		// f.RunOnce(root, &input)
	})
}

func BenchmarkDAGTwoParallelTransits(t *testing.B) {
	root := context.Background()
	t.Run("run successfully", func(t *testing.B) {
		f := NewDAGTwoParallelTransits()
		var input = "test"
		for i := 0; i < 5; i++ {
			var results = f.Execute(root, &input)
			assert.Equal(t, "0test1test", *results)
		}
		var results = f.RunOnce(root, &input)
		assert.Equal(t, "0test1test", *results)
	})
}

var DAGThreeParallelDelayedWorkflowTransits = []*SimpleDAGWorkflowTransit{
	{
		name:           "input",
		channelInputs:  []string{"input"},
		channelOutputs: []string{"t11", "t12", "t13"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			log.Println("input...")
			time.Sleep(time.Second)
			log.Println("input... finished.")
			return a[0], nil
		},
	}, {
		name:           "transit1",
		channelInputs:  []string{"t11"},
		channelOutputs: []string{"t21"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			log.Println("transit1...")
			time.Sleep(time.Second)
			log.Println("transit1... finished.")
			return a[0], nil
		},
	}, {
		name:           "transit2",
		channelInputs:  []string{"t12"},
		channelOutputs: []string{"t22"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			log.Println("transit2...")
			time.Sleep(time.Second)
			log.Println("transit2... finished.")
			return a[0], nil
		},
	}, {
		name:           "transit3",
		channelInputs:  []string{"t13"},
		channelOutputs: []string{"t23"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			log.Println("transit3...")
			//time.Sleep(time.Second)
			log.Println("transit3... finished.")
			return a[0], nil
		},
	}, {
		name:           "transit",
		channelInputs:  []string{"t21", "t22", "t23"},
		channelOutputs: []string{"output"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			var r string
			for i, c := range a {
				r = r + strconv.Itoa(i) + c.(string)
			}
			return r, nil
		},
	},
}

var DAGOneStraightPipeline = []*SimpleDAGWorkflowTransit{
	{
		name:           "input",
		channelInputs:  []string{"input"},
		channelOutputs: []string{"t11"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			log.Println("input...")
			time.Sleep(time.Second)
			log.Println("input... finished.")
			return a[0], nil
		},
	}, {
		name:           "transit1",
		channelInputs:  []string{"t11"},
		channelOutputs: []string{"output"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			log.Println("transit1...")
			time.Sleep(time.Second)
			log.Println("transit1... finished.")
			return a[0], nil
		},
	},
}

func NewDAGOneStraightPipeline() *SimpleDAG1 {
	f := SimpleDAG1{
		SimpleDAG: *NewSimpleDAG[string, string](),
	}
	f.InitChannels("input", "t11", "output")
	f.InitWorkflow("input", "output")
	return &f
}

func TestSimpleDAGOneStaightPipeline(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	root := context.Background()
	t.Run("normal case", func(t *testing.T) {
		f := NewDAGOneStraightPipeline()
		f.AttachWorkflowTransit(DAGOneStraightPipeline...)
		var input = "test"
		var results = f.RunOnce(root, &input)
		assert.Equal(t, "test", *results)
	})
	t.Run("error case", func(t *testing.T) {
		f := NewDAGOneStraightPipeline()
		transits := DAGOneStraightPipeline
		transits[1].worker = func(ctx context.Context, a ...any) (any, error) {
			log.Println("transit1...")
			time.Sleep(time.Second)
			return nil, errors.New("error(s) occurred")
		}
		f.AttachWorkflowTransit(transits...)
		var input = "test"
		var results = f.RunOnce(root, &input)
		assert.Nil(t, results)
	})
}

// NewDAGThreeParallelDelayedTransits defines a workflow.
func NewDAGThreeParallelDelayedTransits() *SimpleDAG1 {
	f := SimpleDAG1{
		SimpleDAG: *NewSimpleDAG[string, string](),
	}
	f.InitChannels("input", "t11", "t12", "t13", "t21", "t22", "t23", "output")
	//   input                   t11               t21              output
	// ---------> input ----+--------> transit1 --------> transit ----------> output
	//                      |    t12               t22      ^
	//                      +--------> transit2 ------------+
	//                      |    t13               t23      ^
	//                      +--------> transit3 ------------+
	f.InitWorkflow("input", "output")
	return &f
}

func NewMultipleParallelTransitNodesWorkflow(total int) *SimpleDAG1 {
	if total < 1 {
		return nil
	}
	f := SimpleDAG1{
		SimpleDAG: *NewSimpleDAG[string, string](),
	}
	f.InitChannels("input", "output")
	channelInputs := make([]string, total)
	channelOutputs := make([]string, total)
	for i := 0; i < total; i++ {
		err := f.AttachChannels(fmt.Sprintf("t%05d1", i), fmt.Sprintf("t%05d2", i))
		if err != nil {
			return nil
		}
		channelInputs[i] = fmt.Sprintf("t%05d2", i)
		channelOutputs[i] = fmt.Sprintf("t%05d1", i)
	}
	f.InitWorkflow("input", "output")
	for i := 0; i < total; i++ {
		f.AttachWorkflowTransit(NewSimpleDAGWorkflowTransit(
			fmt.Sprintf("transit%05d", i),
			[]string{fmt.Sprintf("t%05d1", i)},
			[]string{fmt.Sprintf("t%05d2", i)},
			func(ctx context.Context, a ...any) (any, error) {
				return a[0], nil
			},
		))
	}
	f.AttachWorkflowTransit(NewSimpleDAGWorkflowTransit(
		"input", []string{"input"}, channelOutputs, func(ctx context.Context, a ...any) (any, error) {
			return a[0], nil
		}), NewSimpleDAGWorkflowTransit(
		"output", channelInputs, []string{"output"}, func(ctx context.Context, a ...any) (any, error) {
			var r string
			for i, c := range a {
				r = r + strconv.Itoa(i) + c.(string)
			}
			return r, nil
		}))
	return &f
}

func TestSimpleDAGContext_Cancel(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
	root := context.Background()
	t.Run("normal case", func(t *testing.T) {
		f := NewDAGThreeParallelDelayedTransits()
		f.AttachWorkflowTransit(DAGThreeParallelDelayedWorkflowTransits...)
		var input = "test"
		var results = f.RunOnce(root, &input)
		assert.Equal(t, "0test1test2test", *results)
	})
	// This unit test is used to find out whether there will be data race issue in the channel map when the transit node reports an error.
	t.Run("error case", func(t *testing.T) {
		f := NewDAGThreeParallelDelayedTransits()
		transits := DAGThreeParallelDelayedWorkflowTransits
		transits[1] = &SimpleDAGWorkflowTransit{
			name:           "transit1",
			channelInputs:  []string{"t11"},
			channelOutputs: []string{"t21"},
			worker: func(ctx context.Context, a ...any) (any, error) {
				log.Println("transit1...")
				time.Sleep(time.Second)
				return nil, errors.New("transit1 reports error(s)")
			},
		}
		f.AttachWorkflowTransit(transits...)
		var input = "test"
		var results = f.RunOnce(root, &input)
		assert.Nil(t, results)
	})
	t.Run("normal case with 10000 parallel transit nodes", func(t *testing.T) {
		f := NewMultipleParallelTransitNodesWorkflow(10000)
		var input = "test"
		var results = f.RunOnce(root, &input)
		t.Log(*results)
	})
	t.Run("error case in one of 10000 parallel transit nodes", func(t *testing.T) {

	})
}

func TestRedundantChannelsError_Error(t *testing.T) {
	t.Run("multi names", func(t *testing.T) {
		err := RedundantChannelsError{channels: []string{"a", "b"}}
		assert.Equal(t, "Redundant channelInputs: a, b", err.Error())
	})
}

func BenchmarkMultipleParallelTransitNodesWorkflow(t *testing.B) {
	root := context.Background()
	t.Run("10000 parallel transit nodes", func(t *testing.B) {
		f := NewMultipleParallelTransitNodesWorkflow(10000)
		var input = "test"
		f.RunOnce(root, &input)
	})
	t.Run(" concat 10000 times", func(t *testing.B) {
		var input = ""
		for i := 0; i < 10000; i++ {
			input += strconv.Itoa(i) + "test"
		}
	})
}

func TestNestedWorkflow(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
	root := context.Background()
	f := NewDAGThreeParallelDelayedTransits()
	transits := DAGThreeParallelDelayedWorkflowTransits
	transits[1] = &SimpleDAGWorkflowTransit{
		name:           "transit1",
		channelInputs:  []string{"t11"},
		channelOutputs: []string{"t21"},
		worker: func(ctx context.Context, a ...any) (any, error) {
			log.Println("transit1...")
			time.Sleep(time.Second)
			return nil, errors.New("transit1 reports error(s)")
		},
	}
	f.AttachWorkflowTransit(transits...)
	var input = "test"
	var results = f.RunOnce(root, &input)
	assert.Nil(t, results)
}

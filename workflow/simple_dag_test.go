package workflow

import (
	"context"
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
		worker: func(a ...any) (any, error) {
			return a[0], nil
		},
	}, &SimpleDAGWorkflowTransit{
		name:           "output",
		channelInputs:  []string{"t11"},
		channelOutputs: []string{"output"},
		worker: func(a ...any) (any, error) {
			return 0, nil
		},
	})
	input := "input"
	assert.Nil(t, f.RunOnce(context.Background(), &input))
}

// NewDAGTwoParallelTransits defines a workflow.
func NewDAGTwoParallelTransits() *SimpleDAG1 {
	f := SimpleDAG1{
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
			//var e = SimpleDAGValueTypeError{actual: new(int), expect: new(string)}
			//var t = SimpleDAGValueTypeError{actual: new(int), expect: new(string)}
			//log.Println(errors.Is(&e, &t))
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
		// f.RunOnce(&input)
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

var DAGThreeParallerlDelayedWorkflowTransits = []*SimpleDAGWorkflowTransit{
	{
		name:           "input",
		channelInputs:  []string{"input"},
		channelOutputs: []string{"t11", "t12", "t13"},
		worker: func(a ...any) (any, error) {
			log.Println("input...")
			time.Sleep(time.Second)
			log.Println("input... finished.")
			return a[0], nil
		},
	}, {
		name:           "transit1",
		channelInputs:  []string{"t11"},
		channelOutputs: []string{"t21"},
		worker: func(a ...any) (any, error) {
			log.Println("transit1...")
			time.Sleep(time.Second)
			log.Println("transit1... finished.")
			return a[0], nil
		},
	}, {
		name:           "transit2",
		channelInputs:  []string{"t12"},
		channelOutputs: []string{"t22"},
		worker: func(a ...any) (any, error) {
			log.Println("transit2...")
			time.Sleep(time.Second)
			log.Println("transit2... finished.")
			return a[0], nil
		},
	}, {
		name:           "transit3",
		channelInputs:  []string{"t13"},
		channelOutputs: []string{"t23"},
		worker: func(a ...any) (any, error) {
			log.Println("transit3...")
			//time.Sleep(time.Second)
			log.Println("transit3... finished.")
			return a[0], nil
		},
	}, {
		name:           "transit",
		channelInputs:  []string{"t21", "t22", "t23"},
		channelOutputs: []string{"output"},
		worker: func(a ...any) (any, error) {
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
		worker: func(a ...any) (any, error) {
			log.Println("input...")
			time.Sleep(time.Second)
			log.Println("input... finished.")
			return a[0], nil
		},
	}, {
		name:           "transit1",
		channelInputs:  []string{"t11"},
		channelOutputs: []string{"output"},
		worker: func(a ...any) (any, error) {
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
	//t.Run("error case", func(t *testing.T) {
	//	f := NewDAGOneStraightPipeline()
	//	transits := DAGOneStraightPipeline
	//	transits[1].worker = func(a ...any) (any, error) {
	//		log.Println("transit1...")
	//		time.Sleep(time.Second)
	//		return nil, errors.New("error(s) occurred")
	//	}
	//	f.AttachWorkflowTransit(transits...)
	//	var input = "test"
	//	var results = f.RunOnce(root, &input)
	//	assert.Nil(t, results)
	//})
}

// NewDAGThreeParallelDelayedTransits defines a workflow.
func NewDAGThreeParallelDelayedTransits() *SimpleDAG1 {
	f := SimpleDAG1{
		SimpleDAG: *NewSimpleDAG[string, string](),
	}
	f.InitChannels("input", "t11", "t12", "t13", "t21", "t22", "t23", "output")
	//         input               t11              output
	// input ----+----> transit1 -------> transit ----------> output
	//           |                 t12       ^
	//           +----> transit2 ------------+
	//           |                 t13       ^
	//           +----> transit3 ------------+
	f.InitWorkflow("input", "output")
	return &f
}

func TestSimpleDAGContext_Cancel(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	root := context.Background()
	t.Run("normal case", func(t *testing.T) {
		f := NewDAGThreeParallelDelayedTransits()
		f.AttachWorkflowTransit(DAGThreeParallerlDelayedWorkflowTransits...)
		var input = "test"
		var results = f.RunOnce(root, &input)
		assert.Equal(t, "0test1test2test", *results)
	})
	//t.Run("error case", func(t *testing.T) {
	//	f := NewDAGThreeParallelDelayedTransits()
	//	transits := DAGThreeParallerlDelayedWorkflowTransits
	//	transits[1] = &SimpleDAGWorkflowTransit{
	//		name:           "transit1",
	//		channelInputs:  []string{"t11"},
	//		channelOutputs: []string{"t21"},
	//		worker: func(a ...any) (any, error) {
	//			log.Println("transit1...")
	//			return nil, errors.New("error(s) occurred")
	//			//time.Sleep(time.Second)
	//			//log.Println("transit1... finished.")
	//			//return a[0], nil
	//		},
	//	}
	//	var input = "test"
	//	var results = f.RunOnce(root, &input)
	//	log.Println(results)
	//})
}

/*
 * Copyright (c) 2023 - 2024 vistart.
 */

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
		SimpleDAG: *NewSimpleDAGWithLogger[string, string](NewSimpleDAGJSONLogger()),
	}
	assert.False(t, f.Exists("input"))
	assert.False(t, f.Exists("output"))
	f.InitChannels("input", "output")
	assert.True(t, f.Exists("input"))
	assert.True(t, f.Exists("output"))
}

func TestSimpleDAGValueTypeError_Error(t *testing.T) {
	f := SimpleDAG1{
		SimpleDAG: *NewSimpleDAGWithLogger[string, string](NewSimpleDAGJSONLogger()),
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
			return 0.1, nil
		},
	})
	input := "input"
	assert.Nil(t, f.Execute(context.Background(), &input))
	assert.Nil(t, f.RunOnce(context.Background(), &input))
}

// NewDAGTwoParallelTransitsWithLogger defines a workflow with logger.
func NewDAGTwoParallelTransitsWithLogger() *SimpleDAG1 {
	f := SimpleDAG1{
		SimpleDAG: *NewSimpleDAGWithLogger[string, string](NewSimpleDAGJSONLogger()),
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
			//var e = ErrSimpleDAGValueType{actual: new(int), expect: new(string)}
			//var t = ErrSimpleDAGValueType{actual: new(int), expect: new(string)}
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
			//var e = ErrSimpleDAGValueType{actual: new(int), expect: new(string)}
			//var t = ErrSimpleDAGValueType{actual: new(int), expect: new(string)}
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
		f := NewDAGTwoParallelTransitsWithLogger()
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
		f := NewDAGTwoParallelTransitsWithLogger()
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
		SimpleDAG: *NewSimpleDAGWithLogger[string, string](NewSimpleDAGJSONLogger()),
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
		SimpleDAG: *NewSimpleDAGWithLogger[string, string](NewSimpleDAGJSONLogger()),
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
		err := ErrRedundantChannels{channels: []string{"a", "b"}}
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

func TestDAGThreeParallelDelayedWorkflowTransits(t *testing.T) {
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

func TestMultiDifferentTimeConsumingTasks(t *testing.T) {
	f := NewSimpleDAGWithLogger[int, int](NewSimpleDAGJSONLogger())
	f.InitChannels("input", "t11", "output")
	//   input             t11               output
	// ---------> input ----+----> transit ---------->
	worker := func(ctx context.Context, a ...any) (any, error) {
		time.Sleep(time.Duration(a[0].(int)) * time.Second)
		return a[0], nil
	}
	transits := []*SimpleDAGWorkflowTransit{
		&SimpleDAGWorkflowTransit{
			name:           "transit",
			channelInputs:  []string{"input"},
			channelOutputs: []string{"t11"},
			worker:         worker,
		}, &SimpleDAGWorkflowTransit{
			name:           "transit",
			channelInputs:  []string{"t11"},
			channelOutputs: []string{"output"},
			worker:         worker,
		},
	}
	f.InitWorkflow("input", "output", transits...)
	t.Run("1s per task", func(t *testing.T) {
		input := 1
		output := f.Execute(context.Background(), &input)
		assert.Equal(t, 1, *output)
	})
	t.Run("2s per task", func(t *testing.T) {
		input := 2
		output := f.Execute(context.Background(), &input)
		assert.Equal(t, 2, *output)
	})
	// Bad case: Actual execution results cannot be predicted.
	//t.Run("1s and 2s per task", func(t *testing.T) {
	//	// This case demonstrates that when the same workflow executes two tasks that take different times in a row,
	//	// the order will be reversed.
	//	input1 := 2
	//	input2 := 1
	//	output1 := new(int)
	//	output2 := new(int)
	//	signal1 := make(chan struct{})
	//	signal2 := make(chan struct{})
	//	signal1to2 := make(chan struct{}) // make sure that thread 2 being started after thread 1.
	//	// The two tasks are executed one after another.
	//	// Task 1 takes 2 seconds
	//	go func() {
	//		t.Log("thread 1 starting at", time.Now())
	//		go func() {
	//			output1 = f.Execute(context.Background(), &input1)
	//			t.Log("thread 1 ended at", time.Now())
	//			signal1 <- struct{}{}
	//		}()
	//		signal1to2 <- struct{}{}
	//	}()
	//	// Task 2 takes 4 seconds
	//	go func() {
	//		<-signal1to2
	//		t.Log("thread 2 starting at", time.Now())
	//		output2 = f.Execute(context.Background(), &input2)
	//		t.Log("thread 2 ended at", time.Now())
	//		signal2 <- struct{}{}
	//	}()
	//	<-signal1
	//	<-signal2
	//	assert.Equal(t, 1, *output1)
	//	assert.Equal(t, 2, *output2)
	//})

	// If you want to execute multiple identical workflows in a short period of time
	// without unpredictable data transfer order, please instantiate a new workflow before each execution.
	f1 := NewSimpleDAGWithLogger[int, int](NewSimpleDAGJSONLogger())
	f1.InitChannels("input", "t11", "output")
	f1.InitWorkflow("input", "output", transits...)
	t.Run("1s and 2s per task in different workflow", func(t *testing.T) {
		input1 := 2
		input2 := 1
		output1 := new(int)
		output2 := new(int)
		signal1 := make(chan struct{})
		signal2 := make(chan struct{})
		signal1to2 := make(chan struct{}) // make sure that thread 2 being started after thread 1.
		go func() {
			t.Log("thread 1 starting at", time.Now())
			go func() {
				output1 = f.Execute(context.Background(), &input1)
				t.Log("thread 1 ended at", time.Now())
				signal1 <- struct{}{}
			}()
			signal1to2 <- struct{}{}
		}()
		go func() {
			<-signal1to2
			t.Log("thread 2 starting at", time.Now())
			output2 = f1.Execute(context.Background(), &input2)
			t.Log("thread 2 ended at", time.Now())
			signal2 <- struct{}{}
		}()
		<-signal1
		<-signal2
		assert.Equal(t, 2, *output1)
		assert.Equal(t, 1, *output2)
	})
}

func TestNestedWorkflow(t *testing.T) {
	f := NewSimpleDAGWithLogger[int, int](NewSimpleDAGJSONLogger())
	f.InitChannels("input", "output", "t11", "t12")
	worker1 := func(ctx context.Context, a ...any) (any, error) {
		log.Println("started at", time.Now())
		time.Sleep(time.Duration(a[0].(int)) * time.Second)
		log.Println("ended at", time.Now())
		return a[0], nil
	}
	worker2 := func(ctx context.Context, a ...any) (any, error) {
		f1 := NewSimpleDAGWithLogger[int, int](NewSimpleDAGJSONLogger())
		f1.InitChannels("input", "output", "t11")
		channelInputs1 := []string{"input"}
		channelOutputs1 := []string{"t11"}
		channelOutputs2 := []string{"output"}
		transits := []*SimpleDAGWorkflowTransit{
			NewSimpleDAGWorkflowTransit("i:input", channelInputs1, channelOutputs1, worker1),
			NewSimpleDAGWorkflowTransit("i:output", channelOutputs1, channelOutputs2, worker1),
		}
		f1.InitWorkflow("input", "output", transits...)
		input := 1
		output := f1.Execute(context.Background(), &input)
		return *output, nil
	}
	channelInputs1 := []string{"input"}
	channelOutputs1 := []string{"t11"}
	channelOutputs2 := []string{"t12"}
	channelOutputs3 := []string{"output"}
	transits := []*SimpleDAGWorkflowTransit{
		NewSimpleDAGWorkflowTransit("input", channelInputs1, channelOutputs1, worker1),
		NewSimpleDAGWorkflowTransit("transit", channelOutputs1, channelOutputs2, worker2),
		NewSimpleDAGWorkflowTransit("output", channelOutputs2, channelOutputs3, worker1),
	}
	f.InitWorkflow("input", "output", transits...)
	input := 1
	output := f.Execute(context.Background(), &input)
	assert.Equal(t, 1, *output)
}

func TestErrWorkerPanicked_Error(t *testing.T) {
	logger := NewSimpleDAGJSONLogger()
	logger.SetFlags(LDebugEnabled)
	f := NewSimpleDAGWithLogger[int, int](logger)
	f.InitChannels("input", "output", "t11", "t12")
	worker1 := func(ctx context.Context, a ...any) (any, error) {
		log.Println("started at", time.Now())
		time.Sleep(time.Duration(a[0].(int)) * time.Second)
		log.Println("ended at", time.Now())
		return a[0], nil
	}
	worker2 := func(ctx context.Context, a ...any) (any, error) {
		log.Println("started at", time.Now())
		panic("worker2 panicked")
		return a[0], nil
	}
	channelInputs1 := []string{"input"}
	channelOutputs1 := []string{"t11"}
	channelOutputs2 := []string{"t12"}
	channelOutputs3 := []string{"output"}
	transits := []*SimpleDAGWorkflowTransit{
		NewSimpleDAGWorkflowTransit("input", channelInputs1, channelOutputs1, worker1),
		NewSimpleDAGWorkflowTransit("transit", channelOutputs1, channelOutputs2, worker2),
		NewSimpleDAGWorkflowTransit("output", channelOutputs2, channelOutputs3, worker1),
	}
	f.InitWorkflow("input", "output", transits...)
	input := 1
	output := f.Execute(context.Background(), &input)
	assert.Nil(t, output)
}

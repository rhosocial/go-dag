// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package simple

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

func TestSimpleDAGChannel_Exists(t *testing.T) {
	f, _ := NewDAG[string, string](
		WithLoggers[string, string](NewLogger()))
	assert.False(t, f.channels.exists("input"))
	assert.False(t, f.channels.exists("output"))
	//f.InitChannels("input", "output")
	//assert.True(t, f.channels.Exists("input"))
	//assert.True(t, f.channels.Exists("output"))

	f1, _ := NewDAG[string, string](
		WithDefaultChannels[string, string]())
	assert.True(t, f1.channels.exists("input"))
	assert.True(t, f1.channels.exists("output"))
}

func TestSimpleDAGValueTypeError_Error(t *testing.T) {
	logger := NewLogger()
	logger.SetFlags(LDebugEnabled)
	f, _ := NewDAG[string, string](
		WithDefaultChannels[string, string](),
		WithChannels[string, string]("t11"),
		WithTransits[string, string](&Transit{
			name:           "input",
			channelInputs:  []string{"input"},
			channelOutputs: []string{"t11"},
			worker: func(ctx context.Context, a ...any) (any, error) {
				return a[0], nil
			},
		}, &Transit{
			name:           "output",
			channelInputs:  []string{"t11"},
			channelOutputs: []string{"output"},
			worker: func(ctx context.Context, a ...any) (any, error) {
				return 0.1, nil
			},
		}),
		WithLoggers[string, string](logger))
	input := "input"
	assert.Nil(t, f.Execute(context.Background(), &input))
	assert.Nil(t, f.RunOnce(context.Background(), &input))
}

// NewDAGTwoParallelTransitsWithLogger defines a workflow with logger.
func NewDAGTwoParallelTransitsWithLogger() *DAG[string, string] {
	f, _ := NewDAG[string, string](
		//         t:input          t:transit1          t:output
		//c:input ----+----> c:t11 ------------> c:t21 -----+----> c:output
		//            |             t:transit2              ^
		//            +----> c:t12 ------------> c:t22 -----+
		WithChannels[string, string]("input", "t11", "t12", "t21", "t22", "output"),
		WithDefaultChannels[string, string](),
		WithTransits[string, string](&Transit{
			name:           "input",
			channelInputs:  []string{"input"},
			channelOutputs: []string{"t11", "t12"},
			worker: func(ctx context.Context, a ...any) (any, error) {
				return a[0], nil
			},
		}, &Transit{
			name:           "transit1",
			channelInputs:  []string{"t11"},
			channelOutputs: []string{"t21"},
			worker: func(ctx context.Context, a ...any) (any, error) {
				return a[0], nil
			},
		}, &Transit{
			name:           "transit2",
			channelInputs:  []string{"t12"},
			channelOutputs: []string{"t22"},
			worker: func(ctx context.Context, a ...any) (any, error) {
				//var e = ErrValueTypeMismatch{actual: new(int), expect: new(string)}
				//var t = ErrValueTypeMismatch{actual: new(int), expect: new(string)}
				//log.Println(errors.Is(&e, &t))
				return a[0], nil
			},
		}, &Transit{
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
		}),
		WithLoggers[string, string](NewLogger()))
	return f
}

// NewDAGTwoParallelTransits defines a workflow.
func NewDAGTwoParallelTransits() *DAG[string, string] {
	f, _ := NewDAG[string, string](
		//         t:input          t:transit1          t:output
		//c:input ----+----> c:t11 ------------> c:t21 -----+----> c:output
		//            |             t:transit2              ^
		//            +----> c:t12 ------------> c:t22 -----+
		WithChannels[string, string]("input", "t11", "t12", "t21", "t22", "output"),
		WithChannelInput[string, string]("input"),
		WithChannelOutput[string, string]("output"),
		WithTransits[string, string](
			NewTransit("input",
				WithInputs("input"),
				WithOutputs("t11", "t12"),
				WithWorker(func(ctx context.Context, a ...any) (any, error) {
					return a[0], nil
				}),
			),
			NewTransit("input",
				WithInputs("t11"),
				WithOutputs("t21"),
				WithWorker(func(ctx context.Context, a ...any) (any, error) {
					return a[0], nil
				}),
			),
			NewTransit("input",
				WithInputs("t12"),
				WithOutputs("t22"),
				WithWorker(func(ctx context.Context, a ...any) (any, error) {
					return a[0], nil
				}),
			),
			NewTransit("input",
				WithInputs("t21", "t22"),
				WithOutputs("output"),
				WithWorker(func(ctx context.Context, a ...any) (any, error) {
					var r string
					for i, c := range a {
						r = r + strconv.Itoa(i) + c.(string)
					}
					return r, nil
				}),
			),
		),
	)
	return f
}

func TestDAGTwoParallelTransits(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	root := context.Background()
	t.Run("run successfully", func(t *testing.T) {
		f := NewDAGTwoParallelTransitsWithLogger()
		var input = "test"
		for i := 0; i < 5; i++ {
			var results = f.Execute(root, &input)
			assert.NotNil(t, results)
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

var DAGThreeParallelDelayedWorkflowTransits = []*Transit{
	NewTransit("input",
		WithInputs("input"),
		WithOutputs("t11", "t12", "t13"),
		WithWorker(func(ctx context.Context, a ...any) (any, error) {
			log.Println("input...")
			time.Sleep(time.Second)
			log.Println("input... finished.")
			return a[0], nil
		}),
	),
	NewTransit("transit1",
		WithInputs("t11"),
		WithOutputs("t21"),
		WithWorker(func(ctx context.Context, a ...any) (any, error) {
			log.Println("transit1...")
			time.Sleep(time.Second)
			log.Println("transit1... finished.")
			return a[0], nil
		}),
	),
	NewTransit("transit2",
		WithInputs("t12"),
		WithOutputs("t22"),
		WithWorker(func(ctx context.Context, a ...any) (any, error) {
			log.Println("transit2...")
			time.Sleep(time.Second)
			log.Println("transit2... finished.")
			return a[0], nil
		}),
	),
	NewTransit("transit3",
		WithInputs("t13"),
		WithOutputs("t23"),
		WithWorker(func(ctx context.Context, a ...any) (any, error) {
			log.Println("transit3...")
			//time.Sleep(time.Second)
			log.Println("transit3... finished.")
			return a[0], nil
		}),
	), NewTransit("transit",
		WithInputs("t21", "t22", "t23"),
		WithOutputs("output"),
		WithWorker(func(ctx context.Context, a ...any) (any, error) {
			var r string
			for i, c := range a {
				r = r + strconv.Itoa(i) + c.(string)
			}
			return r, nil
		}),
	),
}

var DAGOneStraightPipeline = []*Transit{
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

func NewDAGOneStraightPipeline() *DAG[string, string] {
	f, _ := NewDAG[string, string](
		WithDefaultChannels[string, string](),
		WithChannels[string, string]("t11"),
		WithLoggers[string, string](NewLogger()))
	return f
}

func TestSimpleDAGOneStraightPipeline(t *testing.T) {
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
func NewDAGThreeParallelDelayedTransits() *DAG[string, string] {
	f, err := NewDAG[string, string](
		WithLoggers[string, string](NewLogger()),
		WithChannels[string, string]("input", "t11", "t12", "t13", "t21", "t22", "t23", "output"),
		WithDefaultChannels[string, string](),
	)
	//   input                   t11               t21              output
	// ---------> input ----+--------> transit1 --------> transit ----------> output
	//                      |    t12               t22      ^
	//                      +--------> transit2 ------------+
	//                      |    t13               t23      ^
	//                      +--------> transit3 ------------+
	if err != nil {
		panic(err)
	}
	return f
}

func NewMultipleParallelTransitNodesWorkflow(total int) *DAG[string, string] {
	if total < 1 {
		return nil
	}
	f, err := NewDAG[string, string](
		WithDefaultChannels[string, string]())
	if err != nil {
		panic(err)
	}
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
	for i := 0; i < total; i++ {
		f.AttachWorkflowTransit(NewTransit(
			fmt.Sprintf("transit%05d", i),
			WithInputs(fmt.Sprintf("t%05d1", i)),
			WithOutputs(fmt.Sprintf("t%05d2", i)),
			WithWorker(func(ctx context.Context, a ...any) (any, error) {
				return a[0], nil
			}),
		))
	}
	f.AttachWorkflowTransit(NewTransit(
		"input", WithInputs("input"), WithOutputs(channelOutputs...), WithWorker(func(ctx context.Context, a ...any) (any, error) {
			return a[0], nil
		})), NewTransit(
		"output", WithInputs(channelInputs...), WithOutputs("output"), WithWorker(func(ctx context.Context, a ...any) (any, error) {
			var r string
			for i, c := range a {
				r = r + strconv.Itoa(i) + c.(string)
			}
			return r, nil
		})))
	return f
}

func TestSimpleDAGContext_Cancel(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
	root := context.Background()
	t.Run("normal case", func(t *testing.T) {
		f := NewDAGThreeParallelDelayedTransits()
		assert.NotNil(t, f)
		f.AttachWorkflowTransit(DAGThreeParallelDelayedWorkflowTransits...)
		var input = "test"
		var results = f.RunOnce(root, &input)
		assert.NotNil(t, results)
		assert.Equal(t, "0test1test2test", *results)
	})
	// This unit test is used to find out whether there will be data race issue in the channel map when the transit node reports an error.
	t.Run("error case", func(t *testing.T) {
		f := NewDAGThreeParallelDelayedTransits()
		transits := DAGThreeParallelDelayedWorkflowTransits
		transits[1] = &Transit{
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
		assert.NotNil(t, results)
		t.Log(*results)
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
	transits[1] = &Transit{
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
	//   input             t11               output
	// ---------> input ----+----> transit ---------->
	worker := func(ctx context.Context, a ...any) (any, error) {
		time.Sleep(time.Duration(a[0].(int)) * time.Second)
		return a[0], nil
	}
	transits := []*Transit{
		{
			name:           "transit",
			channelInputs:  []string{"input"},
			channelOutputs: []string{"t11"},
			worker:         worker,
		}, {
			name:           "transit",
			channelInputs:  []string{"t11"},
			channelOutputs: []string{"output"},
			worker:         worker,
		},
	}
	f, _ := NewDAG[int, int](
		WithChannels[int, int]("input", "t11", "output"),
		WithDefaultChannels[int, int](),
		WithTransits[int, int](transits...),
		WithLoggers[int, int](NewLogger()))
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
	f1, _ := NewDAG[int, int](
		WithChannels[int, int]("input", "t11", "output"),
		WithDefaultChannels[int, int](),
		WithTransits[int, int](transits...),
		WithLoggers[int, int](NewLogger()))
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
		assert.NotNil(t, output1)
		assert.NotNil(t, output2)
		assert.Equal(t, 2, *output1)
		assert.Equal(t, 1, *output2)
	})
}

func TestNestedWorkflow(t *testing.T) {
	worker1 := func(ctx context.Context, a ...any) (any, error) {
		log.Println("started at", time.Now())
		time.Sleep(time.Duration(a[0].(int)) * time.Second)
		log.Println("ended at", time.Now())
		return a[0], nil
	}
	worker2 := func(ctx context.Context, a ...any) (any, error) {
		channelInputs1 := []string{"input"}
		channelOutputs1 := []string{"t11"}
		channelOutputs2 := []string{"output"}
		transits := []*Transit{
			NewTransit("i:input", WithInputs(channelInputs1...), WithOutputs(channelOutputs1...), WithWorker(worker1)),
			NewTransit("i:output", WithInputs(channelOutputs1...), WithOutputs(channelOutputs2...), WithWorker(worker1)),
		}
		f1, _ := NewDAG[int, int](
			WithDefaultChannels[int, int](),
			WithChannels[int, int]("t11"),
			WithTransits[int, int](transits...),
			WithLoggers[int, int](NewLogger()))
		input := 1
		output := f1.Execute(ctx, &input)
		return *output, nil
	}
	channelInputs1 := []string{"input"}
	channelOutputs1 := []string{"t11"}
	channelOutputs2 := []string{"t12"}
	channelOutputs3 := []string{"output"}
	transits := []*Transit{
		NewTransit("input",
			WithInputs(channelInputs1...),
			WithOutputs(channelOutputs1...),
			WithWorker(worker1)),
		NewTransit("transit",
			WithInputs(channelOutputs1...),
			WithOutputs(channelOutputs2...),
			WithWorker(worker2)),
		NewTransit("output", WithInputs(channelOutputs2...), WithOutputs(channelOutputs3...), WithWorker(worker1)),
	}
	f, _ := NewDAG[int, int](
		WithDefaultChannels[int, int](),
		WithChannels[int, int]("t11", "t12"),
		WithTransits[int, int](transits...),
		WithLoggers[int, int](NewLogger()))
	input := 1
	output := f.Execute(context.Background(), &input)
	assert.NotNil(t, output)
	assert.Equal(t, 1, *output)
}

func TestErrWorkerPanicked_Error(t *testing.T) {
	logger := NewLogger()
	logger.SetFlags(LDebugEnabled)
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
	transits := []*Transit{
		NewTransit("input",
			WithInputs(channelInputs1...),
			WithOutputs(channelOutputs1...),
			WithWorker(worker1)),
		NewTransit("transit",
			WithInputs(channelOutputs1...),
			WithOutputs(channelOutputs2...),
			WithWorker(worker2)),
		NewTransit("output", WithInputs(channelOutputs2...), WithOutputs(channelOutputs3...), WithWorker(worker1)),
	}
	f, _ := NewDAG[int, int](
		WithChannels[int, int]("t11", "t12"),
		WithTransits[int, int](transits...),
		WithLoggers[int, int](NewLogger()))
	input := 1
	output := f.Execute(context.Background(), &input)
	assert.Nil(t, output)
}

func TestNewDAGByChainingMethodStyle(t *testing.T) {
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
	transits := []*Transit{
		NewTransit("input",
			WithInputs(channelInputs1...),
			WithOutputs(channelOutputs1...),
			WithWorker(worker1)),
		NewTransit("transit",
			WithInputs(channelOutputs1...),
			WithOutputs(channelOutputs2...),
			WithWorker(worker2)),
		NewTransit("output", WithInputs(channelOutputs2...), WithOutputs(channelOutputs3...), WithWorker(worker1)),
	}
	logger := NewLogger()
	logger.SetFlags(LDebugEnabled)
	f, _ := NewDAG[int, int](
		WithChannels[int, int]("input", "output", "t11", "t12"),
		WithChannelInput[int, int]("input"),
		WithChannelOutput[int, int]("output"),
		WithTransits[int, int](transits...),
		WithLoggers[int, int](logger),
	)
	input := 1
	output := f.Execute(context.Background(), &input)
	assert.Nil(t, output)
	time.Sleep(time.Millisecond)
}

func TestCancelWorkflow(t *testing.T) {
	worker := func(ctx context.Context, a ...any) (any, error) {
		log.Println("started at", time.Now())
		time.Sleep(time.Duration(a[0].(int)) * time.Second)
		log.Println("ended at", time.Now())
		return a[0], nil
	}
	channelInputs1 := []string{"input"}
	channelOutputs1 := []string{"t11"}
	channelOutputs2 := []string{"t12"}
	channelOutputs3 := []string{"output"}
	transits := []*Transit{
		NewTransit("input",
			WithInputs(channelInputs1...),
			WithOutputs(channelOutputs1...),
			WithWorker(worker)),
		NewTransit("transit",
			WithInputs(channelOutputs1...),
			WithOutputs(channelOutputs2...),
			WithWorker(worker)),
		NewTransit("output", WithInputs(channelOutputs2...), WithOutputs(channelOutputs3...), WithWorker(worker)),
	}
	logger := NewLogger()
	logger.SetFlags(LDebugEnabled)
	f, _ := NewDAG[int, int](
		WithChannels[int, int]("input", "output", "t11", "t12"),
		WithChannelInput[int, int]("input"),
		WithChannelOutput[int, int]("output"),
		WithTransits[int, int](transits...),
		WithLoggers[int, int](logger),
	)
	input := 1
	t.Run("cancel before run", func(t *testing.T) {
		f.Cancel(errors.New("cancel before run"))
	})
	t.Run("cancel when running", func(t *testing.T) {
		ch1 := make(chan struct{})
		go func() {
			output := f.Execute(context.Background(), &input)
			assert.Nil(t, output)
			ch1 <- struct{}{}
		}()
		go func() {
			time.Sleep(time.Millisecond * 1500)
			f.Cancel(errors.New("canceled"))
		}()
		<-ch1
		time.Sleep(time.Millisecond)
	})
}

func TestCancelWorkflowWithNestedWorkflow(t *testing.T) {
	logger := NewLogger()
	logger.SetFlags(LDebugEnabled)
	worker1 := func(ctx context.Context, a ...any) (any, error) {
		log.Println("started at", time.Now())
		time.Sleep(time.Duration(a[0].(int)) * time.Second)
		log.Println("ended at", time.Now())
		return a[0], nil
	}
	worker2 := func(ctx context.Context, a ...any) (any, error) {
		channelInputs1 := []string{"input"}
		channelOutputs1 := []string{"t11"}
		channelOutputs2 := []string{"output"}
		transits := []*Transit{
			NewTransit("i:input", WithInputs(channelInputs1...), WithOutputs(channelOutputs1...), WithWorker(worker1)),
			NewTransit("i:output", WithInputs(channelOutputs1...), WithOutputs(channelOutputs2...), WithWorker(worker1)),
		}
		f1, _ := NewDAG[int, int](
			WithDefaultChannels[int, int](),
			WithChannels[int, int]("t11"),
			WithTransits[int, int](transits...),
			WithLoggers[int, int](logger))
		input := 1
		output := f1.Execute(ctx, &input)
		if output == nil {
			return nil, nil
		}
		return *output, nil
	}
	channelInputs1 := []string{"input"}
	channelOutputs1 := []string{"t11"}
	channelOutputs2 := []string{"t12"}
	channelOutputs3 := []string{"output"}
	transits := []*Transit{
		NewTransit("input",
			WithInputs(channelInputs1...),
			WithOutputs(channelOutputs1...),
			WithWorker(worker1)),
		NewTransit("transit",
			WithInputs(channelOutputs1...),
			WithOutputs(channelOutputs2...),
			WithWorker(worker2)),
		NewTransit("output", WithInputs(channelOutputs2...), WithOutputs(channelOutputs3...), WithWorker(worker1)),
	}
	f, _ := NewDAG[int, int](
		WithDefaultChannels[int, int](),
		WithChannels[int, int]("t11", "t12"),
		WithTransits[int, int](transits...),
		WithLoggers[int, int](logger))
	input := 1
	t.Run("cancel before run", func(t *testing.T) {
		f.Cancel(errors.New("cancel before run"))
	})
	t.Run("cancel when running", func(t *testing.T) {
		ch1 := make(chan struct{})
		go func() {
			output := f.Execute(context.Background(), &input)
			assert.Nil(t, output)
			ch1 <- struct{}{}
		}()
		go func() {
			time.Sleep(time.Millisecond * 1500)
			f.Cancel(errors.New("canceled"))
		}()
		<-ch1
		time.Sleep(time.Millisecond)
	})
}

func TestListenErrorReported(t *testing.T) {
	logger := NewLogger()
	logger.SetFlags(LDebugEnabled)
	errorCollector := NewErrorCollector()
	go errorCollector.Listen(context.Background())
	worker1 := func(ctx context.Context, a ...any) (any, error) {
		log.Println("started at", time.Now())
		time.Sleep(time.Duration(a[0].(int)) * time.Second)
		log.Println("ended at", time.Now())
		return a[0], nil
	}
	worker2 := func(ctx context.Context, a ...any) (any, error) {
		channelInputs1 := []string{"input"}
		channelOutputs1 := []string{"t11"}
		channelOutputs2 := []string{"output"}
		transits := []*Transit{
			NewTransit("i:input", WithInputs(channelInputs1...), WithOutputs(channelOutputs1...), WithWorker(worker1)),
			NewTransit("i:output", WithInputs(channelOutputs1...), WithOutputs(channelOutputs2...), WithWorker(worker1)),
		}
		f1, _ := NewDAG[int, int](
			WithDefaultChannels[int, int](),
			WithChannels[int, int]("t11"),
			WithTransits[int, int](transits...),
			WithLoggers[int, int](logger, errorCollector))
		input := 1
		output := f1.Execute(ctx, &input)
		if output == nil {
			return nil, nil
		}
		return *output, nil
	}
	channelInputs1 := []string{"input"}
	channelOutputs1 := []string{"t11"}
	channelOutputs2 := []string{"t12"}
	channelOutputs3 := []string{"output"}
	transits := []*Transit{
		NewTransit("input",
			WithInputs(channelInputs1...),
			WithOutputs(channelOutputs1...),
			WithWorker(worker1)),
		NewTransit("transit",
			WithInputs(channelOutputs1...),
			WithOutputs(channelOutputs2...),
			WithWorker(worker2)),
		NewTransit("output", WithInputs(channelOutputs2...), WithOutputs(channelOutputs3...), WithWorker(worker1)),
	}
	f, _ := NewDAG[int, int](
		WithDefaultChannels[int, int](),
		WithChannels[int, int]("t11", "t12"),
		WithTransits[int, int](transits...),
		WithLoggers[int, int](logger, errorCollector))
	input := 1
	t.Run("cancel before run", func(t *testing.T) {
		f.Cancel(errors.New("cancel before run"))
	})
	t.Run("cancel when running", func(t *testing.T) {
		ch1 := make(chan struct{})
		go func() {
			output := f.Execute(context.Background(), &input)
			assert.Nil(t, output)
			ch1 <- struct{}{}
		}()
		go func() {
			time.Sleep(time.Millisecond * 1500)
			f.Cancel(errors.New("canceled"))
		}()
		<-ch1
		time.Sleep(time.Millisecond)

		errors1 := errorCollector.Get()
		assert.Len(t, errors1, 2)
		assert.IsType(t, LogEventTransitCanceled{}, errors1[0])
		assert.IsType(t, LogEventTransitCanceled{}, errors1[1])
		assert.Equal(t, "i:output", errors1[0].(LogEventTransitCanceled).transit.name)
		assert.Equal(t, "output", errors1[1].(LogEventTransitCanceled).transit.name)
	})
	log.Println("finished")
}

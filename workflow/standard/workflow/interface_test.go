// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package workflow

import (
	baseContext "context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/rhosocial/go-dag/workflow/standard/cache"
	"github.com/rhosocial/go-dag/workflow/standard/channel"
	"github.com/rhosocial/go-dag/workflow/standard/context"
	"github.com/rhosocial/go-dag/workflow/standard/transit"
	"github.com/stretchr/testify/assert"
)

type MockWorkflow struct {
	Interface[struct{}, struct{}]
	RunAsyncInterface[struct{}, struct{}]
}

func (m *MockWorkflow) BuildWorkflowInput(ctx Context, result any, inputs ...string) error {
	return nil
}

func (m *MockWorkflow) BuildWorkflowOutput(ctx Context, outputs ...string) (*[]any, error) {
	return nil, nil
}

func (m *MockWorkflow) BuildChannels(ctx Context) {}

func (m *MockWorkflow) BuildWorkflow(ctx Context) {
}

func (m *MockWorkflow) RunAsyncWithContext(execCtx Context, input struct{}) (<-chan struct{}, <-chan error) {
	return make(<-chan struct{}), nil
}

func (m *MockWorkflow) RunAsync(ctx baseContext.Context, input struct{}) (<-chan struct{}, <-chan error) {
	return make(<-chan struct{}), nil
}

func (m *MockWorkflow) RunWithContext(execCtx Context, input struct{}) (*struct{}, error) {
	return nil, nil
}
func (m *MockWorkflow) Run(ctx baseContext.Context, input struct{}) (*struct{}, error) {
	return nil, nil
}

func (m *MockWorkflow) Close() {}

var _ Interface[struct{}, struct{}] = (*MockWorkflow)(nil)

func TestNewWorkflowWithNilGraph(t *testing.T) {
	_, err := NewWorkflow[struct{}, struct{}](WithGraph[struct{}, struct{}](nil))
	assert.ErrorIs(t, err, GraphNilError{})
	assert.Equal(t, err.Error(), "graph is nil")

	_, err = NewWorkflow[struct{}, struct{}]()
	assert.ErrorIs(t, err, GraphNotSpecifiedError{})
	assert.Equal(t, err.Error(), "graph is not specified")
}

var workerSleepMillisecond = func(ctx baseContext.Context, input ...interface{}) (interface{}, error) {
	time.Sleep(time.Millisecond)
	return input[0], nil
}

var workerSleepSecond = func(ctx baseContext.Context, input ...interface{}) (interface{}, error) {
	time.Sleep(time.Second)
	return input[0], nil
}

var workerErr = func(ctx baseContext.Context, input ...interface{}) (interface{}, error) {
	return nil, errors.New("worker error occurred")
}

func BuildGraph(worker channel.WorkerFunc) (transit.GraphInterface, error) {
	return transit.NewGraph("input", "output", transit.WithIntermediateTransits(
		channel.NewTransit("t1", []string{"input"}, []string{"ch11", "ch12"}, worker),
		channel.NewTransit("t21", []string{"ch11"}, []string{"ch21"}, worker),
		channel.NewTransit("t22", []string{"ch12"}, []string{"ch22"}, worker),
		channel.NewTransit("t3", []string{"ch21", "ch22"}, []string{"output"}, worker),
	))
}

func TestNewWorkflowWithGraphAndTopologicalSort(t *testing.T) {
	graph, err := BuildGraph(workerSleepMillisecond)
	if err != nil {
		t.Fatal(err)
		return
	}
	workflow, err := NewWorkflow[struct{}, struct{}](WithGraph[struct{}, struct{}](graph))
	if err != nil {
		t.Fatal(err)
		return
	}

	output, signal := workflow.RunAsync(baseContext.Background(), struct{}{})
	log.Println("workflow run")
	log.Printf("result: %v, err: %v\n", <-output, <-signal)
	log.Println(workflow.graph.TopologicalSort())
}

func TestWorkflowBuildWorkflowInputChannelNotFound(t *testing.T) {
	wf := &Workflow[struct{}, struct{}]{}
	ctx, err := NewContext("input", "output")
	assert.NoError(t, err)

	err = wf.BuildWorkflowInput(ctx, "result", "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "channel nonexistent not found", err.Error())
}

func TestWorkflowBuildWorkflowOutputChannelNotFound(t *testing.T) {
	wf := &Workflow[struct{}, struct{}]{}
	ctx, err := NewContext("input", "output")
	assert.NoError(t, err)

	_, err = wf.BuildWorkflowOutput(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "channel nonexistent not found", err.Error())
}

func BuildCustomContext() (*workflowContext, error) {
	identifier := context.NewIdentifier(nil)
	log.Println(identifier.GetID())
	// Create a new context with cancel function
	customCtx, cancel := baseContext.WithCancelCause(baseContext.Background())
	return NewContext("input", "output", context.WithContext(customCtx, cancel), context.WithIdentifier(identifier))
}

func TestWorkflowClose(t *testing.T) {
	t.Run("two parallel tasks", func(t *testing.T) {
		graph, err := BuildGraph(workerSleepSecond)
		if err != nil {
			t.Fatal(err)
			return
		}
		workflow, err := NewWorkflow[struct{}, struct{}](WithGraph[struct{}, struct{}](graph))
		if err != nil {
			t.Fatal(err)
			return
		}

		ctx1, _ := BuildCustomContext()
		ctx2, _ := BuildCustomContext()
		err1 := make(chan error, 1)
		err2 := make(chan error, 1)
		go func() {
			_, err := workflow.RunWithContext(ctx1, struct{}{})
			err1 <- err
		}()
		go func() {
			_, err := workflow.RunWithContext(ctx2, struct{}{})
			err2 <- err
		}()
		time.Sleep(time.Millisecond)

		_, ok := workflow.ctxMap.Load(ctx1.GetIdentifier().GetID())
		assert.True(t, ok)

		_, ok = workflow.ctxMap.Load(ctx2.GetIdentifier().GetID())
		assert.True(t, ok)
		time.Sleep(time.Millisecond * 10)
		signal := make(chan struct{})
		go func() {
			<-err1
			_, ok = workflow.ctxMap.Load(ctx1.GetIdentifier().GetID())
			assert.False(t, ok)

			<-err2
			_, ok = workflow.ctxMap.Load(ctx2.GetIdentifier().GetID())
			assert.False(t, ok)
			signal <- struct{}{}
		}()
		workflow.Close()
		<-signal
	})

}

func TestWorkflow_BuildChannels(t *testing.T) {
	t.Run("channel not initialized", func(t *testing.T) {
		graph, err := BuildGraph(workerSleepMillisecond)
		if err != nil {
			t.Fatal(err)
			return
		}
		wf, err := NewWorkflow[struct{}, struct{}](WithGraph[struct{}, struct{}](graph))
		if err != nil {
			t.Fatal(err)
			return
		}
		assert.NoError(t, err)

		identifier := context.NewIdentifier(nil)

		// Create a new context with cancel function
		customCtx, cancel := baseContext.WithCancelCause(baseContext.Background())
		ctx, err := NewContext("input", "output", context.WithContext(customCtx, cancel), context.WithIdentifier(identifier))

		ctx.channels = nil

		wf.BuildChannels(ctx)
		assert.NotNil(t, ctx.channels)
	})
}

func TestWorkflow_Run(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		graph, err := BuildGraph(workerSleepMillisecond)
		if err != nil {
			t.Fatal(err)
			return
		}
		workflow, err := NewWorkflow[struct{}, struct{}](WithGraph[struct{}, struct{}](graph))
		if err != nil {
			t.Fatal(err)
			return
		}
		output, err := workflow.Run(baseContext.Background(), struct{}{})
		assert.NoError(t, err)
		assert.Equal(t, struct{}{}, *output)
	})
	t.Run("run with context", func(t *testing.T) {
		graph, err := BuildGraph(workerSleepMillisecond)
		if err != nil {
			t.Fatal(err)
			return
		}
		workflow, err := NewWorkflow[struct{}, struct{}](WithGraph[struct{}, struct{}](graph))
		if err != nil {
			t.Fatal(err)
			return
		}
		ctx, _ := BuildCustomContext()
		t.Log(ctx.GetIdentifier().GetID())
		output, err := workflow.RunWithContext(ctx, struct{}{})
		assert.NoError(t, err)
		assert.Equal(t, struct{}{}, *output)
	})
	t.Run("two parallel tasks, error occurred", func(t *testing.T) {
		graph, err := BuildGraph(workerSleepSecond)
		if err != nil {
			t.Fatal(err)
			return
		}
		workflow, err := NewWorkflow[struct{}, struct{}](WithGraph[struct{}, struct{}](graph))
		if err != nil {
			t.Fatal(err)
			return
		}

		var err1, err2 error
		signal1 := make(chan struct{}, 1)
		signal2 := make(chan struct{}, 1)
		go func(err *error) {
			_, *err = workflow.Run(baseContext.Background(), struct{}{})
			signal1 <- struct{}{}
		}(&err1)
		go func(err *error) {
			_, *err = workflow.Run(baseContext.Background(), struct{}{})
			signal2 <- struct{}{}
		}(&err2)
		time.Sleep(time.Millisecond * 10)
		workflow.Close()
		<-signal1
		assert.ErrorIs(t, err1, baseContext.Canceled)
		<-signal2
		assert.ErrorIs(t, err2, baseContext.Canceled)
	})
	t.Run("two parallel tasks, run with context, error occurred", func(t *testing.T) {
		graph, err := BuildGraph(workerSleepSecond)
		if err != nil {
			t.Fatal(err)
			return
		}
		workflow, err := NewWorkflow[struct{}, struct{}](WithGraph[struct{}, struct{}](graph))
		if err != nil {
			t.Fatal(err)
			return
		}

		ctx1, _ := BuildCustomContext()
		ctx2, _ := BuildCustomContext()
		var err1, err2 error
		signal1 := make(chan struct{}, 1)
		signal2 := make(chan struct{}, 1)
		go func(ctx *workflowContext, err *error) {
			_, *err = workflow.RunWithContext(ctx, struct{}{})
			signal1 <- struct{}{}
		}(ctx1, &err1)
		go func(ctx *workflowContext, err *error) {
			_, *err = workflow.RunWithContext(ctx, struct{}{})
			signal2 <- struct{}{}
		}(ctx2, &err2)
		time.Sleep(time.Millisecond * 10)

		_, ok := workflow.ctxMap.Load(ctx1.GetIdentifier().GetID())
		assert.True(t, ok)

		_, ok = workflow.ctxMap.Load(ctx2.GetIdentifier().GetID())
		assert.True(t, ok)
		time.Sleep(time.Millisecond)
		workflow.Close()
		time.Sleep(time.Millisecond)

		<-signal1
		_, ok = workflow.ctxMap.Load(ctx1.GetIdentifier().GetID())
		assert.False(t, ok)
		assert.ErrorIs(t, err1, baseContext.Canceled)

		<-signal2
		_, ok = workflow.ctxMap.Load(ctx2.GetIdentifier().GetID())
		assert.False(t, ok)
		assert.ErrorIs(t, err2, baseContext.Canceled)
	})
}

func TestWorkflow_processTransit(t *testing.T) {
	t.Run("build output error", func(t *testing.T) {
		workflow, _ := NewWorkflow[struct{}, struct{}]()
		ctx, _ := BuildCustomContext()
		transit := channel.NewTransit("transit", []string{"input"}, []string{"output"}, workerSleepMillisecond)
		workflow.processTransit(ctx, transit.GetName(), transit)
	})
	t.Run("worker error occurred", func(t *testing.T) {
		graph, _ := transit.NewGraph("input", "output", transit.WithIntermediateTransits(
			channel.NewTransit("t1", []string{"input"}, []string{"ch11", "ch12"}, workerSleepMillisecond),
			channel.NewTransit("t21", []string{"ch11"}, []string{"ch21"}, workerSleepMillisecond),
			channel.NewTransit("t22", []string{"ch12"}, []string{"ch22"}, workerErr),
			channel.NewTransit("t3", []string{"ch21", "ch22"}, []string{"output"}, workerSleepMillisecond),
		))
		workflow, _ := NewWorkflow[struct{}, struct{}](WithGraph[struct{}, struct{}](graph))

		_, signal := workflow.RunAsync(baseContext.Background(), struct{}{})
		log.Println("workflow run")
		log.Printf("err: %v\n", <-signal)
		log.Println(workflow.graph.TopologicalSort())
	})
	t.Run("worker not found", func(t *testing.T) {
		workflow, _ := NewWorkflow[struct{}, struct{}]()
		ctx, _ := BuildCustomContext()
		transit := channel.NewTransit("transit", []string{"input"}, []string{"output"}, nil)
		workflow.processTransit(ctx, transit.GetName(), transit)
	})
}

type TestKeyGetterFunc struct {
	fn KeyGetterFunc
	KeyGetter
}

func (t TestKeyGetterFunc) GetKeyGetterFunc() KeyGetterFunc {
	return t.fn
}

type TestKeyGetter struct {
	inputs []any
	cache.KeyGetter
}

func (t TestKeyGetter) GetKey() string {
	return fmt.Sprintf("%v", t.inputs)
}

var keyFn = TestKeyGetterFunc{
	fn: func(a ...any) cache.KeyGetter {
		return TestKeyGetter{inputs: a}
	},
}

func BuildGraphWithCache(worker channel.WorkerFunc) (transit.GraphInterface, error) {
	return transit.NewGraph("input", "output", transit.WithIntermediateTransits(
		channel.NewTransit("t1", []string{"input"}, []string{"ch11", "ch12"}, worker),
		NewTransit("t21", []string{"ch11"}, []string{"ch21"}, worker, TransitWithCacheKeyGetter(keyFn), TransitWithCache(cache.NewMemoryCache())),
		NewTransit("t22", []string{"ch12"}, []string{"ch22"}, worker, TransitWithCacheKeyGetter(keyFn)),
		channel.NewTransit("t3", []string{"ch21", "ch22"}, []string{"output"}, worker),
	))
}

func TestNewWorkflowWithCache(t *testing.T) {
	graph, err := BuildGraphWithCache(workerSleepMillisecond)
	if err != nil {
		t.Fatal(err)
		return
	}
	workflow, err := NewWorkflow[string, string](
		WithGraph[string, string](graph),
	)
	if err != nil {
		t.Fatal(err)
		return
	}

	output, signal := workflow.RunAsync(baseContext.Background(), "transit_with_cache")
	log.Println("workflow run")
	log.Printf("result: %v, err: %v\n", <-output, <-signal)
	log.Println(workflow.graph.TopologicalSort())
}

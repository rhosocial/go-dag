// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package workflow

import (
	baseContext "context"
	"log"
	"testing"
	"time"

	"github.com/rhosocial/go-dag/workflow/standard/channel"
	"github.com/rhosocial/go-dag/workflow/standard/context"
	"github.com/rhosocial/go-dag/workflow/standard/transit"
	"github.com/stretchr/testify/assert"
)

type MockWorkflow struct {
	Interface[struct{}, struct{}]
	RunAsyncInterface[struct{}, struct{}]
}

func (m *MockWorkflow) BuildWorkflowInput(ctx *Context, result any, inputs ...string) error {
	return nil
}

func (m *MockWorkflow) BuildWorkflowOutput(ctx *Context, outputs ...string) (*[]any, error) {
	return nil, nil
}

func (m *MockWorkflow) BuildChannels(ctx *Context) {}

func (m *MockWorkflow) BuildWorkflow(ctx *Context) error {
	return nil
}

func (m *MockWorkflow) RunAsyncWithContext(execCtx *Context, input struct{}, output chan<- struct{}, err chan<- error) {
}

func (m *MockWorkflow) RunAsync(ctx baseContext.Context, input struct{}, output chan<- struct{}, err chan<- error) {
}

func (m *MockWorkflow) RunWithContext(execCtx *Context, input struct{}) (*struct{}, error) {
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

	output := make(chan struct{}, 1)
	signal := make(chan error)
	go workflow.RunAsync(baseContext.Background(), struct{}{}, output, signal)
	log.Println("workflow run")
	log.Printf("result: %v, err: %v\n", <-output, <-signal)
}

func TestWorkflowBuildWorkflowInput(t *testing.T) {
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
		ctx, err := NewContext(
			context.WithContext(customCtx, cancel),
			context.WithIdentifier(identifier),
		)

		ctx.channels = nil

		err = wf.BuildWorkflowInput(ctx, "result", "input1")
		assert.ErrorIs(t, err, ChannelNotInitializedError{})
	})

}

func TestWorkflowBuildWorkflowOutput(t *testing.T) {
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
		ctx, err := NewContext(
			context.WithContext(customCtx, cancel),
			context.WithIdentifier(identifier),
		)

		ctx.channels = nil

		_, err = wf.BuildWorkflowOutput(ctx, "input1")
		assert.ErrorIs(t, err, ChannelNotInitializedError{})
	})
}

func TestWorkflowBuildWorkflowInputChannelNotFound(t *testing.T) {
	wf := &Workflow[struct{}, struct{}]{}
	ctx, err := NewContext()
	assert.NoError(t, err)

	err = wf.BuildWorkflowInput(ctx, "result", "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "channel nonexistent not found", err.Error())
}

func TestWorkflowBuildWorkflowOutputChannelNotFound(t *testing.T) {
	wf := &Workflow[struct{}, struct{}]{}
	ctx, err := NewContext()
	assert.NoError(t, err)

	_, err = wf.BuildWorkflowOutput(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "channel nonexistent not exist", err.Error())
}

func BuildCustomContext() (*Context, error) {
	identifier := context.NewIdentifier(nil)
	log.Println(identifier.GetID())
	// Create a new context with cancel function
	customCtx, cancel := baseContext.WithCancelCause(baseContext.Background())
	return NewContext(
		context.WithContext(customCtx, cancel),
		context.WithIdentifier(identifier),
	)
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

		_, ok := workflow.ctxMap.Load(ctx1.Identifier.GetID())
		assert.True(t, ok)

		_, ok = workflow.ctxMap.Load(ctx2.Identifier.GetID())
		assert.True(t, ok)
		time.Sleep(time.Millisecond)
		signal := make(chan struct{})
		go func() {
			<-err1
			_, ok = workflow.ctxMap.Load(ctx1.Identifier.GetID())
			assert.False(t, ok)

			<-err2
			_, ok = workflow.ctxMap.Load(ctx2.Identifier.GetID())
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
		ctx, err := NewContext(
			context.WithContext(customCtx, cancel),
			context.WithIdentifier(identifier),
		)

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
		t.Log(ctx.Identifier.GetID())
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
		go func(err *error) {
			_, *err = workflow.Run(baseContext.Background(), struct{}{})
		}(&err1)
		go func(err *error) {
			_, *err = workflow.Run(baseContext.Background(), struct{}{})
		}(&err2)
		time.Sleep(time.Millisecond * 10)
		workflow.Close()
		time.Sleep(time.Millisecond)
		assert.ErrorIs(t, err1, baseContext.Canceled)
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
		go func(ctx *Context, err *error) {
			_, *err = workflow.RunWithContext(ctx, struct{}{})
		}(ctx1, &err1)
		go func(ctx *Context, err *error) {
			_, *err = workflow.RunWithContext(ctx, struct{}{})
		}(ctx2, &err2)
		time.Sleep(time.Millisecond * 10)

		_, ok := workflow.ctxMap.Load(ctx1.Identifier.GetID())
		assert.True(t, ok)

		_, ok = workflow.ctxMap.Load(ctx2.Identifier.GetID())
		assert.True(t, ok)
		time.Sleep(time.Millisecond)
		workflow.Close()
		time.Sleep(time.Millisecond)

		_, ok = workflow.ctxMap.Load(ctx1.Identifier.GetID())
		assert.False(t, ok)

		_, ok = workflow.ctxMap.Load(ctx2.Identifier.GetID())
		assert.False(t, ok)
		assert.ErrorIs(t, err1, baseContext.Canceled)
		assert.ErrorIs(t, err2, baseContext.Canceled)
	})
}

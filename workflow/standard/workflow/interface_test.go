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

func (m *MockWorkflow) RunWithContext(execCtx *Context, input struct{}, output chan<- struct{}) error {
	return nil
}

func (m *MockWorkflow) Run(ctx baseContext.Context, input struct{}, output chan<- struct{}) error {
	return nil
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

	output := make(chan struct{})
	go workflow.Run(baseContext.Background(), struct{}{}, output)
	log.Println("workflow run")
	log.Println("result:", <-output)
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

		output1 := make(chan struct{}, 1)
		output2 := make(chan struct{}, 1)
		ctx1, _ := BuildCustomContext()
		ctx2, _ := BuildCustomContext()
		go workflow.RunWithContext(ctx1, struct{}{}, output1)
		go workflow.RunWithContext(ctx2, struct{}{}, output2)
		time.Sleep(time.Millisecond)

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

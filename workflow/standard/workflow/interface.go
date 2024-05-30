// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package workflow implements a flexible workflow system based on directed acyclic graphs (DAGs).
package workflow

import (
	baseContext "context"
	"fmt"
	"sync"

	"github.com/rhosocial/go-dag/workflow/standard/cache"
	"github.com/rhosocial/go-dag/workflow/standard/channel"
	"github.com/rhosocial/go-dag/workflow/standard/context"
	"github.com/rhosocial/go-dag/workflow/standard/transit"
)

// RunAsyncInterface defines asynchronous execution methods for a workflow.
type RunAsyncInterface[TInput, TOutput any] interface {
	// RunAsyncWithContext executes a workflow asynchronously with a custom context, input, and output channel.
	RunAsyncWithContext(execCtx *Context, input TInput, output chan<- TOutput, err chan<- error)
	// RunAsync executes a workflow asynchronously with the given input and sends the result to the output channel.
	RunAsync(ctx baseContext.Context, input TInput, output chan<- TOutput, err chan<- error)
}

// Interface defines the methods required for a workflow.
type Interface[TInput, TOutput any] interface {
	// BuildWorkflowInput initializes the input channels for the workflow.
	BuildWorkflowInput(ctx *Context, result any, inputs ...string) error
	// BuildWorkflowOutput initializes the output channels for the workflow and retrieves the results.
	BuildWorkflowOutput(ctx *Context, outputs ...string) (*[]any, error)
	// BuildWorkflow sets up the workflow channels and goroutines based on the DAG.
	BuildWorkflow(ctx *Context) error
	// BuildChannels initializes the channels required for the workflow.
	BuildChannels(ctx *Context)
	// RunWithContext executes the workflow with a custom context, input, and output channel.
	RunWithContext(execCtx *Context, input TInput) (*TOutput, error)
	// Run executes the workflow with the given input and sends the result to the output channel.
	Run(ctx baseContext.Context, input TInput) (*TOutput, error)
	// Close cleans up the workflow by cancelling all contexts.
	Close()
}

// Workflow represents a workflow with input type TInput and output type TOutput.
type Workflow[TInput, TOutput any] struct {
	Interface[TInput, TOutput]
	RunAsyncInterface[TInput, TOutput]
	cache  cache.Interface        // Cache for storing intermediate results.
	graph  transit.GraphInterface // The DAG representing the workflow.
	ctxMap sync.Map               // Map for storing execution contexts.
}

// BuildWorkflowInput initializes the input channels for the workflow.
func (wf *Workflow[TInput, TOutput]) BuildWorkflowInput(ctx *Context, result any, inputs ...string) error {
	if ctx.channels == nil {
		return ChannelNotInitializedError{}
	}
	var chs = make([]chan any, len(inputs))
	{
		for i := 0; i < len(inputs); i++ {
			i := i
			if ch, existed := ctx.channels[inputs[i]]; existed {
				chs[i] = ch
			} else {
				return fmt.Errorf("channel %s not found", inputs[i])
			}
		}
	}
	for i := 0; i < len(inputs); i++ {
		i := i
		go func() {
			for {
				select {
				case <-ctx.GetContext().Done():
					return
				case chs[i] <- result:
					// d.Log(ctx, LogEventChannelInputReady{LogEventChannelReady{value: result, name: inputs[i]}})
					return
				default:
				}
			}
		}()
	}
	return nil
}

// BuildWorkflowOutput initializes the output channels for the workflow and retrieves the results.
func (wf *Workflow[TInput, TOutput]) BuildWorkflowOutput(ctx *Context, outputs ...string) (*[]any, error) {
	if ctx.channels == nil {
		return nil, ChannelNotInitializedError{}
	}
	count := len(outputs)
	var results = make([]any, count)
	var chs = make([]chan any, count)
	{
		for i := 0; i < count; i++ {
			i := i
			if ch, existed := ctx.channels[outputs[i]]; existed {
				chs[i] = ch
			} else {
				return nil, fmt.Errorf("channel %s not exist", outputs[i])
			}
		}
	}
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		i := i
		go func() {
			defer wg.Done()
			for { // always check the done notification.
				select {
				case results[i] = <-chs[i]:
					return
				case <-ctx.GetContext().Done():
					return // return immediately if done received and no longer wait for the channel.
				}
			}
		}()
	}
	wg.Wait()
	return &results, nil
}

// BuildWorkflow sets up the channels and goroutines for the workflow based on the DAG.
func (wf *Workflow[TInput, TOutput]) BuildWorkflow(ctx *Context) error {
	transits := wf.graph.GetTransit()

	// Start processing transits
	for name, t := range transits {
		ctx.wg.Add(1)
		go func(name string, transit channel.Transit) {
			defer ctx.wg.Done()

			worker := transit.GetWorker()
			if worker == nil {
				ctx.Cancel(channel.NewWorkerNotSpecifiedError(name))
				return
			}
			inputs, err := wf.BuildWorkflowOutput(ctx, transit.GetIncoming()...)
			if err != nil {
				ctx.Cancel(err)
				return
			}

			// Execute the worker function with the inputs and context
			result, err := worker(ctx.GetContext(), *inputs...)
			if err != nil {
				// If error occurs, propagate the error through the context
				ctx.Cancel(err)
				return
			}

			// Pass the result to the next stage in the workflow
			err = wf.BuildWorkflowInput(ctx, result, transit.GetOutgoing()...)
			if err != nil {
				ctx.Cancel(err)
				return
			}
		}(name, t)
	}

	return nil
}

// BuildChannels initializes the channels required for the workflow.
func (wf *Workflow[TInput, TOutput]) BuildChannels(ctx *Context) {
	if ctx.channels == nil {
		ctx.channels = make(map[string]chan any)
	}
	for _, t := range wf.graph.GetTransit() {
		for _, incoming := range t.GetIncoming() {
			name := incoming
			if _, existed := ctx.channels[name]; !existed {
				ctx.channels[name] = make(chan any)
			}
		}
		for _, outgoing := range t.GetOutgoing() {
			name := outgoing
			if _, existed := ctx.channels[name]; !existed {
				ctx.channels[name] = make(chan any)
			}
		}
	}
}

// RunAsyncWithContext executes a workflow asynchronously with a custom context, input, and output channel.
func (wf *Workflow[TInput, TOutput]) RunAsyncWithContext(execCtx *Context, input TInput, output chan<- TOutput, err chan<- error) {
	// Build channels for the workflow
	wf.BuildChannels(execCtx)
	// This method does not return an error.
	_ = wf.BuildWorkflow(execCtx)
	// Store the context in the map
	wf.ctxMap.Store(execCtx.Identifier.GetID(), execCtx)
	// Clean up context map
	// TODO: Before clearing, the current execution context needs to be moved to the specified location.
	defer wf.ctxMap.Delete(execCtx.Identifier.GetID())

	// Handle output
	var results *TOutput
	signal := make(chan struct{})
	go func(ctx *Context) {
		defer func() {
			signal <- struct{}{}
		}()
		r, _ := wf.BuildWorkflowOutput(execCtx, transit.OutputNodeName)
		select {
		case <-execCtx.GetContext().Done(): // If the end notification has been received, it will exit directly.
			return
		default:
		}
		if ra, ok := (*r)[0].(TOutput); ok {
			results = &ra
		} else {
			results = nil
		}
	}(execCtx)
	wf.BuildWorkflowInput(execCtx, input, transit.InputNodeName)
	<-signal

	// Check if context was cancelled due to an error
	if execCtx.GetContext().Err() != nil {
		err <- execCtx.GetContext().Err()
		return
	}

	output <- *results
	err <- nil
}

// RunWithContext executes the workflow with the given input and sends the result to the output channel using a custom context.
func (wf *Workflow[TInput, TOutput]) RunWithContext(execCtx *Context, input TInput) (*TOutput, error) {
	output := make(chan TOutput, 1)
	signal := make(chan error)
	go wf.RunAsyncWithContext(execCtx, input, output, signal)
	err := <-signal
	if err != nil {
		return nil, err
	}
	result := <-output
	return &result, nil
}

// RunAsync executes a workflow asynchronously with the given input and sends the result to the output channel.
func (wf *Workflow[TInput, TOutput]) RunAsync(ctx baseContext.Context, input TInput, output chan<- TOutput, err chan<- error) {
	// Generate a new identifier for this run
	identifier := context.NewIdentifier(nil)

	// Create a new context with cancel function
	customCtx, cancel := baseContext.WithCancelCause(ctx)
	execCtx, _ := NewContext(
		context.WithContext(customCtx, cancel),
		context.WithIdentifier(identifier),
	)
	// Since the WithContext and WithIdentifier methods do not return errors, the error judgment is ignored here.
	wf.RunAsyncWithContext(execCtx, input, output, err)
}

// Run executes the workflow with the given input and sends the result to the output channel.
func (wf *Workflow[TInput, TOutput]) Run(ctx baseContext.Context, input TInput) (*TOutput, error) {
	output := make(chan TOutput, 1)
	signal := make(chan error)
	go wf.RunAsync(ctx, input, output, signal)
	err := <-signal
	if err != nil {
		return nil, err
	}
	result := <-output
	return &result, nil
}

// Close cleans up the workflow by cancelling all contexts.
func (wf *Workflow[TInput, TOutput]) Close() {
	wf.ctxMap.Range(func(key, value interface{}) bool {
		if ctx, ok := value.(*Context); ok {
			ctx.Cancel(&context.CancelError{Message: "workflow closed"})
		}
		return true
	})
}

// Option defines a function type for configuring a workflow.
type Option[TInput, TOutput any] func(workflow *Workflow[TInput, TOutput]) error

// NewWorkflow instantiates a workflow by specifying the types of input and output (TInput and TOutput).
//
// options is used for instantiating a workflow.
// You must call the WithGraph() method to specify an execution graph for the workflow,
// otherwise a GraphNotSpecifiedError will be reported.
func NewWorkflow[TInput, TOutput any](options ...Option[TInput, TOutput]) (*Workflow[TInput, TOutput], error) {
	workflow := &Workflow[TInput, TOutput]{}
	for _, option := range options {
		err := option(workflow)
		if err != nil {
			return nil, err
		}
	}
	if workflow.graph == nil {
		return nil, GraphNotSpecifiedError{}
	}
	return workflow, nil
}

// WithCache specifies a public caching component for the workflow.
// cache can be nil, meaning caching is not enabled.
func WithCache[TInput, TOutput any](cache cache.Interface) Option[TInput, TOutput] {
	return func(workflow *Workflow[TInput, TOutput]) error {
		workflow.cache = cache
		return nil
	}
}

// WithGraph specifies an execution graph (DAG) for the workflow.
// The execution graph must be instantiated, otherwise a GraphNilError will be reported.
func WithGraph[TInput, TOutput any](graph transit.GraphInterface) Option[TInput, TOutput] {
	return func(workflow *Workflow[TInput, TOutput]) error {
		if graph == nil {
			return GraphNilError{}
		}
		workflow.graph = graph
		return nil
	}
}

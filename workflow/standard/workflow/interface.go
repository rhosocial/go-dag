// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package workflow implements a flexible workflow system based on directed acyclic graphs (DAGs).
package workflow

import (
	baseContext "context"
	"reflect"
	"sync"

	"github.com/rhosocial/go-dag/workflow/standard/cache"
	"github.com/rhosocial/go-dag/workflow/standard/channel"
	"github.com/rhosocial/go-dag/workflow/standard/context"
	"github.com/rhosocial/go-dag/workflow/standard/transit"
)

// RunAsyncInterface defines asynchronous execution methods for a workflow.
//
// Note that there methods must be called asynchronously (add the `go` keyword before calling the method),
// otherwise these methods will block subsequent execution.
//
// All methods of this interface involve the output channel and the err channel.
//
// The err channel will be sent the error result immediately after the entire workflow execution ends.
// If there is no error, it will be sent nil.
// When the channel receives error, the content of the output channel can be accessed.
// Therefore, it is strongly recommended to set the output channel to a buffer by default,
// and the err channel to be unbuffered.
// If the received error is not nil, it means that the workflow execution has an error,
// and the output channel may not send content at this time.
// Therefore, you cannot use whether the output channel receives content
// as a basis for judging the end of workflow execution
// unless you can ensure that the workflow execution will not report an error.
type RunAsyncInterface[TInput, TOutput any] interface {
	// RunAsyncWithContext executes a workflow asynchronously with a custom context, input, and output channel.
	//
	// ctx is a context constructed using NewContext.
	// input is the input content. output is the output channel, and err is the error channel.
	RunAsyncWithContext(execCtx Context, input TInput, output chan<- TOutput, err chan<- error)
	// RunAsync executes a workflow asynchronously with the given input and sends the result to the output channel.
	//
	// ctx is the context.Context of the internal package.
	// If you require a customized context, use RunAsyncWithContext and pass in a custom constructed context.
	// input is the input content. output is the output channel, and err is the error channel.
	RunAsync(ctx baseContext.Context, input TInput, output chan<- TOutput, err chan<- error)
}

// Interface defines the methods required for a workflow.
type Interface[TInput, TOutput any] interface {
	// BuildWorkflowInput initializes the input channels for the workflow.
	BuildWorkflowInput(ctx Context, result any, inputs ...string) error
	// BuildWorkflowOutput initializes the output channels for the workflow and retrieves the results.
	BuildWorkflowOutput(ctx Context, outputs ...string) (*[]any, error)
	// BuildWorkflow sets up the workflow channels and goroutines based on the DAG.
	BuildWorkflow(ctx Context)
	// BuildChannels initializes the channels required for the workflow.
	BuildChannels(ctx Context)
	// RunWithContext executes the workflow with a custom context, input, and output channel.
	RunWithContext(execCtx Context, input TInput) (*TOutput, error)
	// Run executes the workflow with the given input and sends the result to the output channel.
	Run(ctx baseContext.Context, input TInput) (*TOutput, error)
	// Close cleans up the workflow by cancelling all contexts.
	Close()
}

// Workflow represents a workflow with input type TInput and output type TOutput.
type Workflow[TInput, TOutput any] struct {
	Interface[TInput, TOutput]
	RunAsyncInterface[TInput, TOutput]
	graph  transit.GraphInterface // The DAG representing the workflow.
	ctxMap sync.Map               // Map for storing execution contexts.
}

// BuildWorkflowInput initializes the input channels for the workflow.
func (wf *Workflow[TInput, TOutput]) BuildWorkflowInput(ctx Context, result any, inputs ...string) error {
	var chs = make([]chan<- any, len(inputs))
	{
		for i := 0; i < len(inputs); i++ {
			i := i
			if ch, existed := ctx.GetChannel(inputs[i]); existed {
				chs[i] = ch
			} else {
				return ChannelNotFoundError{name: inputs[i]}
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
func (wf *Workflow[TInput, TOutput]) BuildWorkflowOutput(ctx Context, outputs ...string) (*[]any, error) {
	count := len(outputs)
	var results = make([]any, count)
	var chs = make([]<-chan any, count)
	{
		for i := 0; i < count; i++ {
			i := i
			if ch, existed := ctx.GetChannel(outputs[i]); existed {
				chs[i] = ch
			} else {
				return nil, ChannelNotFoundError{name: outputs[i]}
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

// processTransit processes a single transit in the workflow.
func (wf *Workflow[TInput, TOutput]) processTransit(ctx Context, name string, transit channel.Transit) {
	worker := transit.GetWorker()
	if worker == nil {
		ctx.Cancel(channel.NewWorkerNotSpecifiedError(name))
		return
	}
	var err error
	inputs, err := wf.BuildWorkflowOutput(ctx, transit.GetIncoming()...)
	if err != nil {
		ctx.Cancel(err)
		return
	}

	cacheMissed := true
	transitCacheEnabled := CheckIfImplements(transit, (*KeyGetter)(nil)) &&
		CheckIfImplements(transit, (*CacheInterface)(nil)) && transit.(CacheInterface) != nil &&
		transit.(CacheInterface).GetCacheEnabled()
	// Log point: transit cache enabled
	var cacheInterface CacheInterface
	var key cache.KeyGetter
	var result any
	if transitCacheEnabled {
		keyGetter := transit.(KeyGetter).GetKeyGetterFunc()
		if keyGetter != nil {
			key = keyGetter(*inputs...)
			var errCache error
			cacheInterface = transit.(CacheInterface)
			result, errCache = cacheInterface.GetCache().Get(key)
			cacheMissed = errCache != nil
		}
		// Log point: cache missed
	}

	// Execute the worker function with the inputs and context
	if cacheMissed {
		result, err = worker(ctx.GetContext(), *inputs...)
		if err != nil {
			// If error occurs, propagate the error through the context
			ctx.Cancel(err)
			return
		}
		if transitCacheEnabled {
			_ = cacheInterface.GetCache().Set(key, result)
		}
	}

	// Pass the result to the next stage in the workflow
	err = wf.BuildWorkflowInput(ctx, result, transit.GetOutgoing()...)
	if err != nil {
		ctx.Cancel(err)
		return
	}
}

// BuildWorkflow sets up the channels and goroutines for the workflow based on the DAG.
func (wf *Workflow[TInput, TOutput]) BuildWorkflow(ctx Context) {
	transits := wf.graph.GetTransit()

	// Start processing transits
	for name, t := range transits {
		go wf.processTransit(ctx, name, t)
	}

	return
}

// BuildChannels initializes the channels required for the workflow.
func (wf *Workflow[TInput, TOutput]) BuildChannels(ctx Context) {
	ctx.BuildChannels()
	for _, t := range wf.graph.GetTransit() {
		ctx.AppendChannels(t.GetIncoming()...)
		ctx.AppendChannels(t.GetOutgoing()...)
	}
}

// RunAsyncWithContext executes a workflow asynchronously with a custom context, input, and output channel.
func (wf *Workflow[TInput, TOutput]) RunAsyncWithContext(execCtx Context, input TInput, output chan<- TOutput, err chan<- error) {
	// Build channels for the workflow
	wf.BuildChannels(execCtx)
	// This method does not return an error.
	wf.BuildWorkflow(execCtx)
	// Store the context in the map
	wf.ctxMap.Store(execCtx.GetIdentifier().GetID(), execCtx)
	// Clean up context map
	// TODO: Before clearing, the current execution context needs to be moved to the specified location.
	defer wf.ctxMap.Delete(execCtx.GetIdentifier().GetID())

	// Handle output
	var results *TOutput
	signal := make(chan struct{})
	go func(ctx Context) {
		defer func() {
			signal <- struct{}{}
		}()
		r, err := wf.BuildWorkflowOutput(execCtx, execCtx.GetChannelOutput())
		if err != nil {
			execCtx.Cancel(err)
			return
		}
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
	// TODO: to handle the error returned by BuildWorkflowInput.
	_ = wf.BuildWorkflowInput(execCtx, input, execCtx.GetChannelInput())
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
func (wf *Workflow[TInput, TOutput]) RunWithContext(execCtx Context, input TInput) (*TOutput, error) {
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
//
// Note: "input" and "output" are fixed to the names of the workflow's input and output channels.
// If not, please use RunAsyncWithContext instead.
func (wf *Workflow[TInput, TOutput]) RunAsync(ctx baseContext.Context, input TInput, output chan<- TOutput, err chan<- error) {
	// Generate a new identifier for this run
	identifier := context.NewIdentifier(nil)

	// Create a new context with cancel function
	customCtx, cancel := baseContext.WithCancelCause(ctx)
	execCtx, _ := NewContext("input", "output", context.WithContext(customCtx, cancel), context.WithIdentifier(identifier))
	// Since the WithContext and WithIdentifier methods do not return errors, the error judgment is ignored here.
	wf.RunAsyncWithContext(execCtx, input, output, err)
}

// Run executes the workflow with the given input and sends the result to the output channel.
//
// Note: "input" and "output" are fixed to the names of the workflow's input and output channels.
// If not, please use RunWithContext instead.
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
		if ctx, ok := value.(Context); ok {
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
//
// By default, caching is not enabled. If you need to enable it, please pass WithCacheEnabled(true).
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

// CheckIfImplements checks if param implements the targetType interface.
func CheckIfImplements(param any, targetType any) bool {
	paramType := reflect.TypeOf(param)
	if paramType == nil {
		return false
	}
	targetTypeType := reflect.TypeOf(targetType).Elem()
	return paramType.Implements(targetTypeType)
}

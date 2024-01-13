// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

/*
Package simple implements a simple workflow that is executed according to a specified directed acyclic graph.
*/
package simple

import (
	"context"
	"sync"
)

// Transit defines the transit node of a directed acyclic graph.
type Transit struct {
	// The name of the transit node.
	//
	// Although it is not required to be unique, it is not recommended to duplicate the name or leave it blank,
	// and it is not recommended to use `input` or `output` because it may be a reserved name in the future.
	name string

	// channelInputs defines the input channel names required by this transit node.
	//
	// All channelInputs are listened to when building a workflow. Each channel is unbuffered.
	// The transit node will only be executed when all channelInputs have passed in values.
	// Therefore, you need to ensure that each channel must have data coming in, otherwise it will always be blocked.
	channelInputs []string

	// channelOutputs defines the output channel names required by this transit node.
	//
	// The results of the execution of the transit node will be sequentially sent to the designated output channel,
	// so that the subsequent node can continue to execute. Therefore, you need to ensure that each channel must
	// have and only one receiver, otherwise it will always be blocked.
	channelOutputs []string

	// worker represents the working method of this transit.
	//
	// The first parameter is context.Context, which is used to receive the context derived from the superior to
	// the current task. All contexts within a worker can receive done notifications for this context.
	// If your worker is time-consuming, it should be able to receive incoming done notifications from the context.
	//
	// The next parameter list is the parameters passed in by the channel defined by channelInputs. You need to check
	// the type of each parameter passed in yourself.
	//
	// The first return value is the execution result of the current worker. If it is a complex value, it is recommended
	// to pass a pointer. It will be sent to each channel defined by channelOutputs in order.
	// You need to ensure the correctness of the parameter and return value types by yourself, otherwise it will panic.
	//
	// worker can return error, but error are not passed on to subsequent transits.
	// If the worker returns an error other than nil, the workflow will be canceled directly.
	//
	// It is strongly discouraged to panic() in the worker, as this will cause the process to abort and
	// cannot be recovered.
	worker func(context.Context, ...any) (any, error)
}

type DAGInitInterface interface {
	// AttachChannels attaches additional channel names. All channelInputs are unbuffered.
	//
	// If the channel channelInputs already exists, the channel will be recreated.
	AttachChannels(channels ...string) error

	// AttachWorkflowTransit attaches additional transit node of workflow.
	AttachWorkflowTransit(...*Transit)
}

// DAGInterface defines a set of methods that a simple DAG should implement.
//
// TInput represents the input data type, and TOutput represents the output data type.
type DAGInterface[TInput, TOutput any] interface {
	DAGInitInterface
	// BuildWorkflow builds the workflow.
	//
	// Input and output channelInputs must be specified, otherwise error is returned.
	BuildWorkflow(ctx context.Context) error

	// BuildWorkflowInput builds channel(s) for inputting the execution result of the current node
	// to the subsequent node(s).
	//
	// Among the parameters, result is the content to be input to subsequent node(s).
	// inputs is the channelInputs of the channelInputs that receive the content.
	// You should check that the channel exists before calling this method, otherwise it will panic.
	BuildWorkflowInput(ctx context.Context, result any, inputs ...string)

	// BuildWorkflowOutput builds a set of channelInputs for receiving execution results of predecessor nodes.
	//
	// outputs is the channelInputs of the channelInputs that content will send to in order.
	// Only when all channelInputs specified by outputs have received data will the result be returned.
	// That is, as long as there is a channel that has not received data, the method will always be blocked.
	// The returned result is an array, the number of array elements is consistent with outputs,
	// and the results of its elements are in the order specified by outputs.
	// You should check that the channel exists before calling this method, otherwise it will panic.
	//
	// In particular, when ctx receives the end notification, it will return immediately,
	// but the content returned at this time is incomplete.
	// Therefore, you need to check the state of ctx immediately after receiving the return value of this method.
	// If the ctx status is `Done()`, the data returned by this method is not available.
	BuildWorkflowOutput(ctx context.Context, outputs ...string) *[]any

	// CloseWorkflow closes all channelInputs after workflow execution.
	//
	// Note: if you want to ensure that the workflow is only executed once,
	// execution of this method is deferred immediately after BuildWorkflow.
	CloseWorkflow()

	// Execute builds the workflow and executes it.
	// All channelInputs will not be closed after this method is executed, that is, you can execute it again and again.
	// But this method can only be executed once at a time.
	// If this method is called before the last call is completed, the call will be blocked until
	// the last execution ends. That is, pipelined execution is not supported.
	// If you want to execute multiple times simultaneously without blocking,
	// you should newly instantiate multiple DAG instances and execute them separately.
	Execute(ctx context.Context, input *TInput) *TOutput

	// RunOnce executes the workflow only once. All channelInputs are closed after execution.
	// If you want to re-execute, you need to re-create the workflow instance. (call NewDAG)
	RunOnce(ctx context.Context, input *TInput) *TOutput

	// Cancel an executing workflow. If there are no workflows executing, there will be no impact.
	Cancel(cause error)

	// Log a log.
	Log(ctx context.Context, events ...LogEventInterface)
}

type ChannelsInterface interface {
	exists(name string) bool
	get(name string) (chan any, error)
	add(names ...string) error
}

// Channels defines the channelInputs used by this directed acyclic graph.
type Channels struct {
	// channels stores all channels of this directed acyclic graph. The key of the map is the channel channels.
	channels   map[string]chan any
	muChannels sync.RWMutex
	// channelInput defines the channel used for input. Currently only a single input channel is supported.
	channelInput string
	// channelOutput defines the channel used for output. Currently only a single output channel is supported.
	channelOutput string

	ChannelsInterface
}

// NewDAGChannels instantiates a map of channels.
func NewDAGChannels() *Channels {
	return &Channels{
		channels: make(map[string]chan any),
	}
}

func (d *Channels) exists(name string) bool {
	if d == nil || d.channels == nil {
		return false
	}
	if _, existed := d.channels[name]; existed {
		return true
	}
	return false
}

// get the specified name of channel.
// Before calling, you must ensure that the channel list has been initialized
// and the specified channel exists, otherwise an error will be reported.
func (d *Channels) get(name string) (chan any, error) {
	if d == nil {
		return nil, ErrChannelNotInitialized
	}
	// d.muChannels.RLock()
	//defer d.muChannels.RUnlock()
	if d.channels == nil {
		return nil, ErrChannelNotInitialized
	}
	if _, existed := d.channels[name]; !existed {
		return nil, ErrChannelNotExist{name: name}
	}
	return d.channels[name], nil
}

// add channels.
// Note that the channel name to be added cannot already exist. Otherwise, `ErrChannelNameExisted` will be returned.
func (d *Channels) add(names ...string) error {
	if d == nil {
		return ErrChannelNotInitialized
	}
	d.muChannels.Lock()
	defer d.muChannels.Unlock()
	for _, v := range names {
		v := v
		if d.exists(v) {
			return ErrChannelNameExisted{name: v}
		}
	}
	// Can only be added after all checks are passed.
	for _, v := range names {
		v := v
		d.channels[v] = make(chan any)
	}
	return nil
}

type ContextInterface interface {
	Cancel(cause error)
}

// Context defines the context, and the context must support cancellation reasons.
// When a node in the workflow reports an error during execution, the cancel function will be called to notify all nodes
// in the workflow to cancel the execution.
type Context struct {
	context context.Context
	cancel  context.CancelCauseFunc
	ContextInterface
}

// Cancel the execution.
// `nil` means no reason, but it is strongly recommended not to do this.
func (d *Context) Cancel(cause error) {
	d.cancel(cause)
}

type Transits struct {
	muTransits sync.RWMutex
	transits   []*Transit
}

type Loggers struct {
	muLoggers sync.RWMutex
	loggers   []LoggerInterface
}

func (d *Loggers) Log(ctx context.Context, events ...LogEventInterface) {
	d.muLoggers.RLock()
	defer d.muLoggers.RUnlock()
	if d.loggers == nil || len(d.loggers) == 0 {
		return
	}
	for _, logger := range d.loggers {
		if logger == nil {
			continue
		}
		logger.Log(ctx, events...)
	}
}

// DAG defines a generic directed acyclic graph of proposals.
// When you use it, you need to specify the input data type and output data type.
//
// Note that the input and output data types of the transit node are not mandatory,
// you need to verify it yourself.
type DAG[TInput, TOutput any] struct {
	muChannels sync.RWMutex
	channels   *Channels
	muContext  sync.RWMutex
	context    *Context
	muTransits sync.RWMutex
	transits   *Transits
	muLoggers  sync.RWMutex
	loggers    *Loggers
	DAGInterface[TInput, TOutput]
}

func (d *DAG[TInput, TOutput]) Log(ctx context.Context, events ...LogEventInterface) {
	if events == nil || len(events) == 0 {
		return
	}
	d.muLoggers.RLock()
	defer d.muLoggers.RUnlock()
	if d.loggers == nil {
		return
	}
	d.loggers.Log(ctx, events...)
}

// AttachChannels attaches the channels to the workflow.
// The parameter `channels` is the name list of the channels to be initialized for the first time, and the order of
// each element is irrelevant. If you don't pass any parameter, it will have no effect.
func (d *DAG[TInput, TOutput]) AttachChannels(channels ...string) error {
	if d.channels == nil {
		d.channels = NewDAGChannels()
	}
	if channels == nil || len(channels) == 0 {
		return nil
	}
	return d.channels.add(channels...)
}

// AttachWorkflowTransit attaches the transits to the workflow.
// The parameter transits represents a list of transit definitions, and the order of each element is irrelevant.
func (d *DAG[TInput, TOutput]) AttachWorkflowTransit(transits ...*Transit) {
	if d.transits == nil {
		d.transits = &Transits{}
	}
	d.transits.transits = append(d.transits.transits, transits...)
}

// BuildWorkflowInput feeds the result to each input channel in turn.
//
// This function does not currently need to consider the termination of the superior context notification.
func (d *DAG[TInput, TOutput]) BuildWorkflowInput(ctx context.Context, result any, inputs ...string) {
	d.muChannels.RLock()
	defer d.muChannels.RUnlock()
	var chs = make([]chan any, len(inputs))
	{
		d.muChannels.RLock()
		defer d.muChannels.RUnlock()

		if d.channels == nil {
			return
		}
		for i := 0; i < len(inputs); i++ {
			i := i
			if ch, err := d.channels.get(inputs[i]); err == nil {
				chs[i] = ch
			}
		}
	}
	for i := 0; i < len(inputs); i++ {
		i := i
		go func() {
			if chs[i] != nil {
				chs[i] <- result
			}
		}()
	}
}

// BuildWorkflowOutput will wait for all channels specified by the current transit to output content,
// and then return the content together.
//
// If there is a channel that does not return in time, the current function will remain blocked.
// If you do not want to be blocked when calling the current method, you should use an asynchronous style.
// If a done notification is received from the superior context.Context, all coroutines will return immediately.
// At this point, `results[i]` that have not received data are still `nil`.
func (d *DAG[TInput, TOutput]) BuildWorkflowOutput(ctx context.Context, outputs ...string) *[]any {
	var count = len(outputs)
	var results = make([]any, count)
	var chs = make([]chan any, count)
	{
		d.muChannels.RLock()
		defer d.muChannels.RUnlock()

		if d.channels == nil {
			return &results
		}
		for i := 0; i < count; i++ {
			i := i
			if ch, err := d.channels.get(outputs[i]); err == nil {
				chs[i] = ch
			}
		}
	}
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		i := i
		go func() {
			defer wg.Done()
			if chs[i] == nil {
				return
			}
			for { // always check the done notification.
				select {
				case results[i] = <-chs[i]:
					return
				case <-ctx.Done():
					return // return immediately if done received and no longer wait for the channel.
				}
			}
		}()
	}
	wg.Wait()
	return &results
}

// BuildWorkflow is used to build the entire workflow.
//
// Note that before building, the input and output channels must be prepared, otherwise an exception will be returned:
// - ErrChannelNotInitialized if channels is nil
// - ErrChannelInputEmpty if channelInput is empty
// - ErrChannelOutputEmpty if channelOutput is empty
//
// If the transits is empty, do nothing and return nil directly.
// Otherwise, it is determined whether the input and output channel names mentioned in each transit are defined
// in the channel. As long as one channel does not exist, an error will be reported: ErrChannelNotExist.
//
// After the check passes, the workflow is built according to the following rules:
// - Call BuildWorkflowOutput to wait for the result of previous transit.
// - Running the worker of transit.
// - Call BuildWorkflowInput with the result of the worker.
func (d *DAG[TInput, TOutput]) BuildWorkflow(ctx context.Context) error {
	d.muChannels.RLock()
	defer d.muChannels.RUnlock()
	if d.channels == nil || d.channels.channels == nil {
		return ErrChannelNotInitialized
	}

	d.channels.muChannels.RLock()
	defer d.channels.muChannels.RUnlock()
	// Checks the channel Input
	if len(d.channels.channelInput) == 0 {
		return ErrChannelInputEmpty
	}

	// Checks the channel Output
	if len(d.channels.channelOutput) == 0 {
		return ErrChannelOutputEmpty
	}

	d.transits.muTransits.RLock()
	defer d.transits.muTransits.RUnlock()
	// Checks the Transits
	if len(d.transits.transits) == 0 {
		return nil
	}
	for _, t := range d.transits.transits {
		for _, name := range t.channelInputs {
			if _, existed := d.channels.channels[name]; !existed {
				return ErrChannelNotExist{name: name}
			}
		}
		for _, name := range t.channelOutputs {
			if _, existed := d.channels.channels[name]; !existed {
				return ErrChannelNotExist{name: name}
			}
		}
	}

	for _, t := range d.transits.transits {
		go func(ctx context.Context, t *Transit) {
			// Waiting for the results of input channels to be ready.
			// If there is a channel with no output, the current coroutine will be blocked here.
			// If done notification is received, return immediately and no longer wait for the channel.
			results := d.BuildWorkflowOutput(ctx, t.channelInputs...)
			select {
			case <-ctx.Done(): // If the cancellation notification has been received, it will exit directly.
				d.Log(ctx, LogEventTransitCanceled{
					LogEventTransitReportedError: LogEventTransitReportedError{
						LogEventTransit: LogEventTransit{transit: t}}})
				return
			default:
				//log.Println("build workflow:", t.channelInputs, " selected.")
			}
			// The sub-Context is derived here only to prevent other Contexts from being affected when the worker stops
			// actively.
			workerCtx, _ := context.WithCancelCause(ctx)
			var work = func(t *Transit) (any, error) {
				return t.worker(workerCtx, *results...)
			}
			d.Log(ctx, LogEventTransitStart{LogEventTransit: LogEventTransit{transit: t}})
			var result, err = func(t *Transit) (any, error) {
				defer func() {
					if err := recover(); err != nil {
						e := ErrWorkerPanicked{panic: err}
						d.Log(ctx, LogEventTransitWorkerPanicked{
							LogEventTransitReportedError: LogEventTransitReportedError{
								LogEventTransit: LogEventTransit{transit: t},
								LogEventError:   LogEventError{err: e}}})
						d.context.Cancel(&e)
					}
				}()
				return work(t)
			}(t)
			d.Log(ctx, LogEventTransitEnd{LogEventTransit: LogEventTransit{transit: t}})
			if err != nil {
				d.Log(ctx, LogEventTransitReportedError{
					LogEventTransit: LogEventTransit{transit: t},
					LogEventError:   NewLogEventError(err)})
				d.context.Cancel(err)
				return
			}
			d.BuildWorkflowInput(ctx, result, t.channelOutputs...)
		}(ctx, t)
	}
	return nil
}

// CloseWorkflow closes the workflow after execution.
// After this method is executed, all input, output channels and transits will be deleted.
func (d *DAG[TInput, TOutput]) CloseWorkflow() {
	d.muChannels.Lock()
	defer d.muChannels.Unlock()
	d.channels = nil
}

// Execute the workflow.
//
// root is the highest-level context.Context for this execution. Each transit worker will receive its child context.
//
// Note that if two tasks with different durations are executed in a short period of time, there is no guarantee that
// the execution results will be as expected, because the input of the subsequent task may be sent to the channel
// before the previous task. If you want to execute multiple identical workflows in a short period of time
// without unpredictable data transfer order, please instantiate a new workflow before each execution.
func (d *DAG[TInput, TOutput]) Execute(root context.Context, input *TInput) *TOutput {
	// The sub-context is introduced and has a cancellation handler, making it easy to terminate the entire process
	// at any time.
	ctx, cancel := context.WithCancelCause(root)

	// Record the context and cancellation handler, so they can be called at the appropriate time.
	d.muContext.Lock()
	d.context = &Context{context: ctx, cancel: cancel}
	d.muContext.Unlock()

	err := d.BuildWorkflow(ctx)
	if err != nil {
		return nil
	}
	var results *TOutput
	signal := make(chan struct{})
	go func(ctx context.Context) {
		defer func() {
			signal <- struct{}{}
		}()
		r := d.BuildWorkflowOutput(ctx, d.channels.channelOutput)
		select {
		case <-ctx.Done(): // If the end notification has been received, it will exit directly.
			return
		default:
		}
		if ra, ok := (*r)[0].(TOutput); ok {
			results = &ra
		} else {
			var a = new(TOutput)
			var e = ErrValueTypeMismatch{actual: (*r)[0], expect: *a}
			d.Log(ctx, LogEventErrorValueTypeMismatch{err: e})
			results = nil
		}
	}(ctx)
	d.BuildWorkflowInput(ctx, *input, d.channels.channelInput)
	<-signal
	return results
}

// RunOnce executes the workflow only once and destroys it. Calling Execute() or RunOnce() again will report an error.
func (d *DAG[TInput, TOutput]) RunOnce(ctx context.Context, input *TInput) *TOutput {
	defer d.CloseWorkflow()
	return d.Execute(ctx, input)
}

// Cancel an executing workflow. If there are no workflow executing, there will be no impact.
func (d *DAG[TInput, TOutput]) Cancel(cause error) {
	d.muContext.RLock()
	defer d.muContext.RUnlock()
	if d.context == nil {
		return
	}
	d.context.Cancel(cause)
}

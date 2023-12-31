/*
 * Copyright (c) 2023 vistart.
 */

package workflow

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
)

// SimpleDAGWorkflowTransit defines the transit node of a directed acyclic graph.
type SimpleDAGWorkflowTransit struct {
	// The name of the transit node.
	//
	// Although it is not required to be unique, it is not recommended to duplicate the name or leave it blank,
	// and it is not recommended to use `input` or `output` because it may be a reserved name in the future.
	name string

	// channelInputs defines the input channel names required by this transit node.
	//
	// All channelInputs are listened to when building a workflow.
	// The transit node will only be executed when all channelInputs have passed in values.
	// Therefore, at least one channel must be specified.
	//
	// Note: Since each channel is unbuffered, it is strongly recommended to use each channel only once.
	channelInputs []string

	// channelOutputs defines the output channel names required by this transit node.
	//
	// The results of the execution of the transit node will be sequentially sent to the designated output channel,
	// so that the subsequent node can continue to execute. Therefore, at least one channel must be specified.
	//
	// Note: Since each channel is unbuffered, it is strongly recommended to use each channel only once.
	channelOutputs []string

	// worker represents the working method of this node.
	//
	// The first parameter is context.Context, which is used to receive the context derived from the superior to
	// the current task. All contexts within a worker can receive done notifications for this context.
	//
	// worker can return error, but error are not passed on to subsequent nodes.
	// If the worker returns an error other than nil, the workflow will be canceled directly.
	//
	// It is strongly discouraged to panic() in the worker, as this will cause the process to abort and
	// cannot be recovered.
	//
	// Note that the parameter list of the method must correspond to the parameters passed in
	// by the channel defined by channelInputs.
	// The output of this method will be sent to each channel defined by channelOutputs in order.
	//
	// You need to ensure the correctness of the parameter and return value types by yourself, otherwise it will panic.
	worker func(context.Context, ...any) (any, error)

	// isRunning indicates that the worker is working.
	isRunning bool
}

func NewSimpleDAGWorkflowTransit(name string, inputs []string, outputs []string,
	worker func(context.Context, ...any) (any, error)) *SimpleDAGWorkflowTransit {
	return &SimpleDAGWorkflowTransit{name: name, channelInputs: inputs, channelOutputs: outputs, worker: worker}
}

type SimpleDAGInitInterface interface {
	// InitChannels initializes the channelInputs that the workflow should have. All channelInputs are unbuffered.
	//
	// The parameter is the channel channelInputs and cannot be repeated.
	InitChannels(channels ...string)

	// AttachChannels attaches additional channel names. All channelInputs are unbuffered.
	//
	// If the channel channelInputs already exists, the channel will be recreated.
	AttachChannels(channels ...string)

	// InitWorkflow initializes the workflow.
	//
	// Among the parameters, input indicates the channelInputs of the input channel,
	// output indicates the channelInputs of the output channel, and transits indicates the transit node.
	//
	// Before executing this method, the corresponding channel(s) must be prepared.
	// And it must be ensured that all created channelInputs can be used. Otherwise, unforeseen consequences may occur.
	InitWorkflow(input string, output string, transits ...*SimpleDAGWorkflowTransit)

	// AttachWorkflowTransit attaches additional transit node of workflow.
	AttachWorkflowTransit(...*SimpleDAGWorkflowTransit)
}

// SimpleDAGInterface defines a set of methods that a simple DAG should implement.
//
// TInput represents the input data type, and TOutput represents the output data type.
type SimpleDAGInterface[TInput, TOutput any] interface {
	SimpleDAGInitInterface
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
	// you should newly instantiate multiple SimpleDAG instances and execute them separately.
	Execute(ctx context.Context, input *TInput) *TOutput

	// RunOnce executes the workflow only once. All channelInputs are closed after execution.
	// If you want to re-execute, you need to re-create the workflow instance. (call NewSimpleDAG)
	RunOnce(ctx context.Context, input *TInput) *TOutput
}

// SimpleDAGChannel defines the channelInputs used by this directed acyclic graph.
type SimpleDAGChannel struct {
	// channels stores all channels of this directed acyclic graph. The key of the map is the channel channels.
	channels   map[string]chan any
	muChannels sync.RWMutex
	// channelInput defines the channel used for input. Currently only a single input channel is supported.
	channelInput string
	// channelOutput defines the channel used for output. Currently only a single output channel is supported.
	channelOutput string
}

func NewSimpleDAGChannel() *SimpleDAGChannel {
	return &SimpleDAGChannel{
		channels: make(map[string]chan any),
	}
}

func (d *SimpleDAGChannel) exists(name string) bool {
	if d.channels == nil {
		return false
	}
	if _, existed := d.channels[name]; existed {
		return true
	}
	return false
}

// GetChannel gets the specified name of channel.
// Before calling, you must ensure that the channel list has been initialized
// and the specified channel exists, otherwise an error will be reported.
func (d *SimpleDAGChannel) GetChannel(name string) (chan any, error) {
	d.muChannels.RLock()
	defer d.muChannels.RUnlock()
	if d.channels == nil {
		return nil, ErrChannelNotInitialized
	}
	if _, existed := d.channels[name]; !existed {
		return nil, ErrChannelNotExist
	}
	return d.channels[name], nil
}

// Send refers to sending the value to the channel with the specified name.
// Before calling, you must ensure that the channel list has been initialized
// and the specified channel exists, otherwise an error will be reported.
func (d *SimpleDAGChannel) Send(name string, value any) error {
	d.muChannels.RLock()
	defer d.muChannels.RUnlock()
	if d.channels == nil {
		return ErrChannelNotInitialized
	}
	if _, existed := d.channels[name]; !existed {
		return ErrChannelNotExist
	}
	d.channels[name] <- value
	return nil
}

type ErrDAGChannelNameExisted struct {
	name string
	error
}

func (e *ErrDAGChannelNameExisted) Error() string {
	return fmt.Sprintf("the channel[%s] has existed.", e.name)
}

// Add channels.
// Note that the channel name to be added cannot already exist. Otherwise, `ErrDAGChannelNameExisted` will be returned.
func (d *SimpleDAGChannel) Add(names ...string) error {
	d.muChannels.Lock()
	defer d.muChannels.Unlock()
	if d.channels == nil {
		return ErrChannelNotInitialized
	}
	for _, v := range names {
		v := v
		if d.exists(v) {
			return &ErrDAGChannelNameExisted{name: v}
		}
	}
	// Can only be added after all checks are passed.
	for _, v := range names {
		v := v
		d.channels[v] = make(chan any)
	}
	return nil
}

// Exists check if the `name` is existed in channel list.
func (d *SimpleDAGChannel) Exists(name string) bool {
	return d.exists(name)
}

type SimpleDAGContextInterface interface {
	Cancel(cause error)
}

// SimpleDAGContext defines the context, and the context must support cancellation reasons.
// When a node in the workflow reports an error during execution, the cancel function will be called to notify all nodes
// in the workflow to cancel the execution.
type SimpleDAGContext struct {
	context context.Context
	cancel  context.CancelCauseFunc
	SimpleDAGContextInterface
}

// Cancel the execution.
// `nil` means no reason, but it is strongly recommended not to do this.
func (d *SimpleDAGContext) Cancel(cause error) {
	log.Println("canceled:", cause.Error())
	//time.Sleep(time.Second)
	d.cancel(cause)
	//log.Println("canceled finished:", cause.Error())
}

type SimpleDAGWorkflowInterface interface {
	IsRunning() bool
	GetRunningWorkflowNames() []string
}

type SimpleDAGWorkflow struct {
	workflowTransits []*SimpleDAGWorkflowTransit
	SimpleDAGWorkflowInterface
}

//func (d *SimpleDAGWorkflow) IsRunning() bool {
//	return true
//}

//func (d *SimpleDAGWorkflow) GetRunningWorkflowNames() *[]string {
//	results := make([]string, 0)
//	return &results
//}

// SimpleDAG defines a generic directed acyclic graph of proposals.
// When you use it, you need to specify the input data type and output data type.
//
// Note that the input and output data types of the transit node are not mandatory,
// you need to verify it yourself.
type SimpleDAG[TInput, TOutput any] struct {
	logger *log.Logger
	SimpleDAGChannel
	SimpleDAGContext
	SimpleDAGWorkflow
	SimpleDAGInterface[TInput, TOutput]
}

// SimpleDAGValueTypeError defines that the data type output by the node is inconsistent with expectation.
type SimpleDAGValueTypeError struct {
	expect any
	actual any
	error
}

func (e *SimpleDAGValueTypeError) Error() string {
	return fmt.Sprintf("The type of the value [%s] is inconsistent with expectation [%s].",
		reflect.TypeOf(e.actual), reflect.TypeOf(e.expect))
}

// NewSimpleDAG instantiates a workflow.
func NewSimpleDAG[TInput, TOutput any]() *SimpleDAG[TInput, TOutput] {
	return &SimpleDAG[TInput, TOutput]{
		logger: log.Default(),
	}
}

// InitChannels initializes the channels.
// The parameter `channels` is the name list of the channels to be initialized for the first time, and the order of
// each element is irrelevant.
// You can pass no parameters, but it is strongly not recommended unless you know the consequences of doing so.
// Generally, you need to specify at least the input and output channels of the workflow.
func (d *SimpleDAG[TInput, TOutput]) InitChannels(channels ...string) error {
	d.SimpleDAGChannel = *NewSimpleDAGChannel()
	return d.AttachChannels(channels...)
}

// AttachChannels attaches the channels to the workflow.
// The parameter `channels` is the name list of the channels to be initialized for the first time, and the order of
// each element is irrelevant. If you don't pass any parameter, it will have no effect.
func (d *SimpleDAG[TInput, TOutput]) AttachChannels(channels ...string) error {
	if channels == nil || len(channels) == 0 {
		return nil
	}
	return d.SimpleDAGChannel.Add(channels...)
}

// InitWorkflow initializes the workflow.
// Input and output channel names need to be specified first. The corresponding channel must already exist.
// Next is the transit list. If this parameter is not passed in, there will be no transit.
func (d *SimpleDAG[TInput, TOutput]) InitWorkflow(input string, output string, transits ...*SimpleDAGWorkflowTransit) {
	d.channelInput = input
	d.channelOutput = output
	lenTransits := len(transits)
	if lenTransits == 0 {
		return
	}
	d.workflowTransits = make([]*SimpleDAGWorkflowTransit, lenTransits)
	for i, t := range transits {
		d.workflowTransits[i] = t
	}
}

// AttachWorkflowTransit attaches the transits to the workflow.
// The parameter transits represents a list of transit definitions, and the order of each element is irrelevant.
func (d *SimpleDAG[TInput, TOutput]) AttachWorkflowTransit(transits ...*SimpleDAGWorkflowTransit) {
	d.workflowTransits = append(d.workflowTransits, transits...)
}

var ErrChannelNotInitialized = errors.New("the channel map is not initialized")

var ErrChannelNotExist = errors.New("the specified channel does not exist")

// BuildWorkflowInput feeds the result to each input channel in turn.
//
// This function does not currently need to consider the termination of the superior context notification.
func (d *SimpleDAG[TInput, TOutput]) BuildWorkflowInput(ctx context.Context, result any, inputs ...string) {
	for i := 0; i < len(inputs); i++ {
		i := i
		var ch chan any
		var err error
		if ch, err = d.GetChannel(inputs[i]); err != nil {
			return
		}
		go func() {
			ch <- result
		}()
	}
	// Please DO NOT use the for-range statements, as it is caused the data race, use c-style for-loop instead.
	//for _, next := range inputs {
	//	next := next
	//	go func(next string) {
	//		log.Println("BuildWorkflowInput[next]:", next)
	//		// TODO: The following statement will cause data race with the access channel map in "CloseWorkflow()".
	//		d.channels[next] <- result
	//	}(next)
	//}
}

// BuildWorkflowOutput will wait for all channels specified by the current transit to output content,
// and then return the content together.
//
// If there is a channel that does not return in time, the current function will remain blocked.
// If you do not want to be blocked when calling the current method, you should use an asynchronous style.
// If a done notification is received from the superior context.Context, all coroutines will return immediately.
// At this point, `results[i]` that have not received data are still `nil`.
func (d *SimpleDAG[TInput, TOutput]) BuildWorkflowOutput(ctx context.Context, outputs ...string) *[]any {
	var count = len(outputs)
	var results = make([]any, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		i := i
		var ch chan any
		var err error
		if ch, err = d.GetChannel(outputs[i]); err != nil {
			wg.Done()
			continue
		}
		go func() {
			defer wg.Done()
			for { // always check the done notification.
				select {
				case results[i] = <-ch:
					return
				case <-ctx.Done():
					return // return immediately if done received and no longer wait for the channel.
				}
			}
		}()
	}
	// Please DO NOT use the following statements, as it is caused the data race, use c-style for-loop instead.
	//for i, name := range outputs {
	//	go func(ctx context.Context, i int, name string) {
	//		defer wg.Done()
	//		for {
	//			select {
	//			case results[i] = <-d.channels[name]:
	//				return
	//			case <-ctx.Done():
	//				return
	//			}
	//		}
	//	}(ctx, i, name)
	//}
	wg.Wait()
	return &results
}

var ErrChannelInputEmpty = errors.New("the input channel is empty")
var ErrChannelOutputEmpty = errors.New("the output channel is empty")

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
func (d *SimpleDAG[TInput, TOutput]) BuildWorkflow(ctx context.Context) error {
	d.muChannels.RLock()
	defer d.muChannels.RUnlock()
	if d.channels == nil {
		return ErrChannelNotInitialized
	}
	// Build Input
	if len(d.channelInput) == 0 {
		return ErrChannelInputEmpty
	}

	// Build Output
	if len(d.channelOutput) == 0 {
		return ErrChannelOutputEmpty
	}

	// Build Transits
	if len(d.workflowTransits) == 0 {
		return nil
	}
	for _, t := range d.workflowTransits {
		for _, name := range t.channelInputs {
			if _, existed := d.channels[name]; !existed {
				return ErrChannelNotExist
			}
		}
		for _, name := range t.channelOutputs {
			if _, existed := d.channels[name]; !existed {
				return ErrChannelNotExist
			}
		}
	}

	for _, t := range d.workflowTransits {
		go func(ctx context.Context, t *SimpleDAGWorkflowTransit) {
			// Waiting for the results of input channels to be ready.
			// If there is a channel with no output, the current coroutine will be blocked here.
			// If done notification is received, return immediately and no longer wait for the channel.
			results := d.BuildWorkflowOutput(ctx, t.channelInputs...)
			select {
			case <-ctx.Done(): // If the end notification has been received, it will exit directly without notifying the worker to work.
				return
			default:
				//log.Println("build workflow:", t.channelInputs, " selected.")
			}
			// The sub-Context is derived here only to prevent other Contexts from being affected when the worker stops
			// actively.
			workerCtx, _ := context.WithCancelCause(ctx)
			var work = func(t *SimpleDAGWorkflowTransit) (any, error) {
				t.isRunning = true
				defer func(t *SimpleDAGWorkflowTransit) {
					t.isRunning = false
				}(t)
				return t.worker(workerCtx, *results...)
			}
			var result, err = work(t)
			if err != nil {
				d.logger.Printf("worker[%s] error(s) occurred: %s\n", t.name, err.Error())
				d.SimpleDAGContext.Cancel(err)
				d.logger.Printf("worker[%s] notified all any other workers to stop.", t.name)
				return
			}
			d.BuildWorkflowInput(ctx, result, t.channelOutputs...)
		}(ctx, t)
	}
	return nil
}

// CloseWorkflow closes the workflow after execution.
// After this method is executed, all input, output channels and transits will be deleted.
func (d *SimpleDAG[TInput, TOutput]) CloseWorkflow() {
	d.muChannels.Lock()
	defer d.muChannels.Unlock()
	if d.channels == nil {
		return
	}
	if len(d.channelInput) == 0 {
		return
	}
	close(d.channels[d.channelInput])
	if len(d.channelOutput) == 0 {
		return
	}
	//close(d.channels[d.channelOutput])
	if len(d.workflowTransits) == 0 {
		return
	}
	d.channels = nil
}

// RedundantChannelsError indicates that there are unused channelInputs.
type RedundantChannelsError struct {
	channels []string
	error
}

func (e *RedundantChannelsError) Error() string {
	return fmt.Sprintf("Redundant channelInputs: %v", strings.Join(e.channels, ", "))
}

// TransitChannelNonExistError indicates that the channel(s) to be used by the specified node does not exist.
type TransitChannelNonExistError struct {
	transitName    string
	channelInputs  []string
	channelOutputs []string
	error
}

func (e *TransitChannelNonExistError) Error() string {
	return fmt.Sprintf("The specified channel(s) does not exist: input[%v], output[%v]",
		strings.Join(e.channelInputs, ", "), strings.Join(e.channelOutputs, ", "))
}

//func (d *SimpleDAG[TInput, TOutput]) CheckChannelUsedInTransits() error {
//	if d.channels == nil {
//		return ErrChannelNotInitialized
//	}
//	channels := make(map[string]struct{})
//	for key := range d.channels {
//		channels[key] = struct{}{}
//	}
//	for _, value := range d.workflowTransits {
//		for _, i := range value.channelInputs {
//			delete(channels, i)
//		}
//		for _, o := range value.channelOutputs {
//			delete(channels, o)
//		}
//	}
//	if len(channels) > 0 {
//		c := make([]string, len(channels))
//		index := 0
//		for i := range channels {
//			c[index] = i
//			index++
//		}
//		return &RedundantChannelsError{channels: c}
//	}
//	return nil
//}

//func (d *SimpleDAG[TInput, TOutput]) CheckChannelExistsInTransits() error {
//	return nil
//}

//func (d *SimpleDAG[TInput, TOutput]) CheckChannelsAndWorkflows() error {
//	return nil
//}

// Execute the workflow.
//
// root is the highest-level context.Context for this execution. Each transit worker will receive its child context.
func (d *SimpleDAG[TInput, TOutput]) Execute(root context.Context, input *TInput) *TOutput {
	// The sub-context is introduced and has a cancellation handler, making it easy to terminate the entire process
	// at any time.
	ctx, cancel := context.WithCancelCause(root)

	// Record the context and cancellation handler, so they can be called at the appropriate time.
	d.SimpleDAGContext = SimpleDAGContext{context: ctx, cancel: cancel}
	err := d.BuildWorkflow(d.SimpleDAGContext.context)
	if err != nil {
		return nil
	}
	var results *TOutput
	signal := make(chan struct{})
	go func(ctx context.Context) {
		defer func() {
			signal <- struct{}{}
		}()
		r := d.BuildWorkflowOutput(ctx, d.channelOutput)
		select {
		case <-ctx.Done(): // If the end notification has been received, it will exit directly.
			return
		default:
		}
		if ra, ok := (*r)[0].(TOutput); ok {
			results = &ra
		} else {
			var a = new(TOutput)
			var e = SimpleDAGValueTypeError{actual: (*r)[0], expect: *a}
			d.logger.Println(e.Error())
			results = nil
		}
	}(d.SimpleDAGContext.context)
	d.BuildWorkflowInput(d.SimpleDAGContext.context, *input, d.channelInput)
	<-signal
	return results
}

// RunOnce executes the workflow only once and destroys it. Calling Execute() or RunOnce() again will report an error.
func (d *SimpleDAG[TInput, TOutput]) RunOnce(ctx context.Context, input *TInput) *TOutput {
	defer d.CloseWorkflow()
	return d.Execute(ctx, input)
}

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
	// All channels are listened to when building a workflow.
	// The transit node will only be executed when all channels have passed in values.
	// Therefore, at least one channel must be specified.
	channelInputs []string

	// channelOutputs defines the output channel names required by this transit node.
	//
	// The results of the execution of the transit node will be sequentially sent to the designated output channel,
	// so that the subsequent node can continue to execute. Therefore, at least one channel must be specified.
	channelOutputs []string

	// worker represents the working method of this node. worker can return error, but error are not passed on to subsequent nodes.
	// If the worker returns an error other than nil, the workflow will be canceled directly.
	//
	// It is strongly discouraged to panic() in the worker, as this will cause the process to abort and cannot be recovered.
	//
	// Note that the parameter list of the method must correspond to the parameters passed in by the channel defined by channelInputs.
	// The output of this method will be sent to each channel defined by channelOutputs in order.
	//
	// You need to ensure the correctness of the parameter and return value types by yourself, otherwise it will panic.
	worker func(...any) (any, error)

	// isRunning indicates that the worker is working.
	isRunning bool

	// mu indicates that the channelInputs, channelOutputs and worker are being accessed.
	mu sync.RWMutex
}

type SimpleDAGInitInterface interface {
	// InitChannels initializes the channels that the workflow should have. All channels are unbuffered.
	//
	// The parameter is the channel name and cannot be repeated.
	InitChannels(channels ...string)

	// AttachChannels attaches additional channel names. All channels are unbuffered.
	//
	// If the channel name already exists, the channel will be recreated.
	AttachChannels(channels ...string)

	// InitWorkflow initializes the workflow.
	//
	// Among the parameters, input indicates the name of the input channel,
	// output indicates the name of the output channel, and transits indicates the transit node.
	//
	// Before executing this method, the corresponding channel(s) must be prepared.
	// And it must be ensured that all created channels can be used. Otherwise, unforeseen consequences may occur.
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
	// Input and output channels must be specified, otherwise error is returned.
	BuildWorkflow(ctx context.Context) error

	// BuildWorkflowInput builds channel(s) for inputting the execution result of the current node to the subsequent node(s).
	//
	// Among the parameters, result is the content to be input to subsequent node(s).
	// inputs is the name of the channels that receive the content.
	// You should check that the channel exists before calling this method, otherwise it will panic.
	BuildWorkflowInput(ctx context.Context, result any, inputs ...string)

	// BuildWorkflowOutput builds a set of channels for receiving execution results of predecessor nodes.
	//
	// outputs is the name of the channels that content will send to in order.
	// You should check that the channel exists before calling this method, otherwise it will panic.
	BuildWorkflowOutput(ctx context.Context, outputs ...string) *[]any

	// CloseWorkflow closes all channels after workflow execution.
	//
	// Note: if you want to ensure that the workflow is only executed once,
	// execution of this method is deferred immediately after BuildWorkflow.
	CloseWorkflow()

	// Execute builds the workflow and executes it.
	//
	// Note:
	// 1. All channels will not be closed after this method is executed, that is, you can execute it again and again.
	// 2. If the current task execution of a node has not ended, but the next task has ended,
	//    the output of the next task will enter the execution flow of the subsequent task.
	//    Therefore, you need to ensure that the order of execution will not be disordered in the middle.
	//    Otherwise, unexpected results may be obtained.
	Execute(ctx context.Context, input *TInput) *TOutput

	// RunOnce executes the workflow only once. All channels are closed after execution.
	// If you want to re-execute, you need to re-create the workflow instance. (call NewSimpleDAG)
	RunOnce(ctx context.Context, input *TInput) *TOutput
}

type SimpleDAGChannel struct {
	// channels stores all channels of this directed acyclic graph. The key of the map is the channel name.
	channels map[string]chan any
	// channelInput defines the channel used for input. Currently only a single input channel is supported.
	channelInput string
	// channelOutput defines the channel used for output. Currently only a single output channel is supported.
	channelOutput string
}

// SimpleDAG defines a generic directed acyclic graph of proposals.
// When you use it, you need to specify the input data type and output data type.
//
// Note that the input and output data types of the transit node are not mandatory,
// you need to verify it yourself.
type SimpleDAG[TInput, TOutput any] struct {
	workflowTransits []*SimpleDAGWorkflowTransit
	logger           *log.Logger
	SimpleDAGChannel
	SimpleDAGInterface[TInput, TOutput]
}

// SimpleDAGValueTypeError defines that the data type output by the node is inconsistent with expectation.
type SimpleDAGValueTypeError struct {
	expect any
	actual any
}

func (e *SimpleDAGValueTypeError) Error() string {
	return fmt.Sprintf("The type of the value [%s] is inconsistent with expectation [%s].", reflect.TypeOf(e.actual), reflect.TypeOf(e.actual))
}

func NewSimpleDAG[TInput, TOutput any]() *SimpleDAG[TInput, TOutput] {
	return &SimpleDAG[TInput, TOutput]{
		logger: log.Default(),
	}
}

func (d *SimpleDAG[TInput, TOutput]) InitChannels(channels ...string) {
	if channels == nil || len(channels) == 0 {
		return
	}
	d.channels = make(map[string]chan any, len(channels))
	for _, v := range channels {
		v := v
		d.channels[v] = make(chan any)
	}
}

func (d *SimpleDAG[TInput, TOutput]) AttachChannels(channels ...string) {
	if channels == nil || len(channels) == 0 {
		return
	}
	if d.channels == nil || len(d.channels) == 0 {
		d.InitChannels(channels...)
		return
	}
	for _, v := range channels {
		v := v
		d.channels[v] = make(chan any)
	}
}

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

func (d *SimpleDAG[TInput, TOutput]) AttachWorkflowTransit(transits ...*SimpleDAGWorkflowTransit) {
	d.workflowTransits = append(d.workflowTransits, transits...)
}

var ErrChannelNotExist = errors.New("the specified channel does not exist")

func (d *SimpleDAG[TInput, TOutput]) BuildWorkflowInput(ctx context.Context, result any, inputs ...string) {
	for _, next := range inputs {
		next := next
		go func(next string) {
			d.channels[next] <- result
		}(next)
	}
}

func (d *SimpleDAG[TInput, TOutput]) BuildWorkflowOutput(ctx context.Context, outputs ...string) *[]any {
	var count = len(outputs)
	var results = make([]any, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i, name := range outputs {
		go func(ctx context.Context, i int, name string) {
			defer wg.Done()
			flag := false
			for {
				if flag {
					break
				}
				select {
				case results[i] = <-d.channels[name]:
					flag = true
				case <-ctx.Done():
					flag = true
				}
			}
		}(ctx, i, name)
	}
	wg.Wait()
	return &results
}

var ErrChannelInputEmpty = errors.New("the input channel is empty")
var ErrChannelOutputEmpty = errors.New("the output channel is empty")

func (d *SimpleDAG[TInput, TOutput]) BuildWorkflow(ctx context.Context) error {
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
			results := d.BuildWorkflowOutput(ctx, t.channelInputs...)
			select {
			case <-ctx.Done():
				return
			default:
			}
			var result, err = t.worker(*results...)
			if err != nil {
				d.logger.Printf("worker[%s] error(s) occurred: %s\n", t.name, err.Error())
				return
			}
			// Note: If the execution of this task has not yet ended, but the next task has ended,
			// the output of the next task will enter the execution flow of this subsequent task.
			d.BuildWorkflowInput(ctx, result, t.channelOutputs...)
		}(ctx, t)
	}
	return nil
}

func (d *SimpleDAG[TInput, TOutput]) CloseWorkflow() {
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
	for _, t := range d.workflowTransits {
		for _, c := range t.channelOutputs {
			close(d.channels[c])
		}
	}
}

func (d *SimpleDAG[TInput, TOutput]) Execute(ctx context.Context, input *TInput) *TOutput {
	err := d.BuildWorkflow(ctx)
	if err != nil {
		return nil
	}
	var results *TOutput
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := d.BuildWorkflowOutput(ctx, d.channelOutput)
		if r, ok := (*r)[0].(TOutput); !ok {
			t := new(TOutput)
			var e = SimpleDAGValueTypeError{actual: r, expect: t}
			d.logger.Println(e)
			results = nil
			return
		} else {
			results = &r
		}
	}()
	d.BuildWorkflowInput(ctx, *input, d.channelInput)
	wg.Wait()
	return results
}

func (d *SimpleDAG[TInput, TOutput]) RunOnce(ctx context.Context, input *TInput) *TOutput {
	defer d.CloseWorkflow()
	return d.Execute(ctx, input)
}

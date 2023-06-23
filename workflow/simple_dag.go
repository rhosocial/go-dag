/*
 * Copyright (c) 2023 vistart.
 */

package workflow

import (
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
	//
	// It is strongly discouraged to panic() in the worker, as this will cause the process to abort and cannot be recovered.
	//
	// Note that the parameter list of the method must correspond to the parameters passed in by the channel defined by channelInputs.
	// The output of this method will be sent to each channel defined by channelOutputs in order.
	//
	// You need to ensure the correctness of the parameter and return value types by yourself, otherwise it will panic.
	worker func(...any) (any, error)
}

// SimpleDAGInterface defines a set of methods that a simple DAG should implement.
//
// TInput represents the input data type, and TOutput represents the output data type.
type SimpleDAGInterface[TInput, TOutput any] interface {
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

	// BuildWorkflow builds the workflow.
	//
	// Input and output channels must be specified, otherwise false is returned.
	BuildWorkflow() error

	// BuildWorkflowInput builds channel(s) for inputting the execution result of the current node to the subsequent node(s).
	//
	// Among the parameters, result is the content to be input to subsequent node(s).
	// inputs is the name of the channels that receive the content.
	// You should check that the channel exists before calling this method, otherwise it will panic.
	BuildWorkflowInput(result any, inputs ...string)

	// BuildWorkflowOutput builds a set of channels for receiving execution results of predecessor nodes.
	//
	// outputs is the name of the channels that content will send to in order.
	// You should check that the channel exists before calling this method, otherwise it will panic.
	BuildWorkflowOutput(outputs ...string) *[]any

	// CloseWorkflow closes all channels after workflow execution.
	//
	// Note: if you want to ensure that the workflow is only executed once,
	// execution of this method is deferred immediately after BuildWorkflow.
	CloseWorkflow()

	// Execute builds the workflow and executes it.
	//
	// Note: all channels will not be closed after this method is executed, that is, you can execute it again and again.
	Execute(input *TInput) *TOutput

	// RunOnce executes the workflow only once. All channels are closed after execution.
	// If you want to re-execute, you need to re-create the workflow instance. (call NewSimpleDAG)
	RunOnce(input *TInput) *TOutput
}

type SimpleDAG[TInput, TOutput any] struct {
	channels         map[string]chan any
	channelInput     string
	channelOutput    string
	workflowTransits []*SimpleDAGWorkflowTransit
	logger           *log.Logger
	SimpleDAGInterface[TInput, TOutput]
}

type SimpleDAGValueTypeError struct {
	expect any
	actual any
}

func (e *SimpleDAGValueTypeError) Error() string {
	return fmt.Sprintf("The type of the value [%s] is inconsistent with the target [%s].", reflect.TypeOf(e.actual), reflect.TypeOf(e.actual))
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
	for _, t := range transits {
		d.workflowTransits = append(d.workflowTransits, t)
	}
}

var ErrChannelNotExist = errors.New("the specified channel does not exist")

func (d *SimpleDAG[TInput, TOutput]) BuildWorkflowInput(result any, inputs ...string) {
	for _, next := range inputs {
		next := next
		go func(next string) {
			d.channels[next] <- result
		}(next)
	}
}

func (d *SimpleDAG[TInput, TOutput]) BuildWorkflowOutput(outputs ...string) *[]any {
	var count = len(outputs)
	var results = make([]any, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i, name := range outputs {
		go func(i int, name string) {
			defer wg.Done()
			results[i] = <-d.channels[name]
		}(i, name)
	}
	wg.Wait()
	return &results
}

var ErrChannelInputEmpty = errors.New("the input channel is empty")
var ErrChannelOutputEmpty = errors.New("the output channel is empty")

func (d *SimpleDAG[TInput, TOutput]) BuildWorkflow() error {
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
		go func(t *SimpleDAGWorkflowTransit) {
			results := d.BuildWorkflowOutput(t.channelInputs...)
			var result, err = t.worker(*results...)
			if err != nil {
				d.logger.Printf("worker[%s] error(s) occurred: %s\n", t.name, err.Error())
			}
			d.BuildWorkflowInput(result, t.channelOutputs...)
		}(t)
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

func (d *SimpleDAG[TInput, TOutput]) Execute(input *TInput) *TOutput {
	err := d.BuildWorkflow()
	if err != nil {
		return nil
	}
	var results *TOutput
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := d.BuildWorkflowOutput(d.channelOutput)
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
	d.BuildWorkflowInput(*input, d.channelInput)
	wg.Wait()
	return results
}

func (d *SimpleDAG[TInput, TOutput]) RunOnce(input *TInput) *TOutput {
	defer d.CloseWorkflow()
	return d.Execute(input)
}

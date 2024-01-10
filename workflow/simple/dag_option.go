// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

/*
Package simple implements a simple workflow that is executed according to a specified directed acyclic graph.
*/
package simple

import "context"

type Option[TInput, TOutput any] func(d *DAG[TInput, TOutput]) error

// NewDAG instantiates a workflow.
//
// The new instance does not come with a logger. If you want to specify a logger, use the NewDAGWithLogger method.
func NewDAG[TInput, TOutput any](options ...Option[TInput, TOutput]) *DAG[TInput, TOutput] {
	dag := &DAG[TInput, TOutput]{}
	for _, option := range options {
		option(dag)
	}
	return dag
}

// WithChannels specifies channel names for the workflow.
//
// Note that names cannot be repeated, and repeated names are counted once.
// All specified names must be used in the transit and may appear only once on the input and once on the output.
// At least two channel names must be specified, for input and output.
func WithChannels[TInput, TOutput any](names ...string) Option[TInput, TOutput] {
	return func(d *DAG[TInput, TOutput]) error {
		if d.channels == nil {
			d.channels = NewDAGChannels()
		}
		return d.channels.Add(names...)
	}
}

// WithChannelInput specifies the input channel for the entire workflow.
//
// When the workflow is executed, input data is injected from this channel.
// After the input data is injected into the channel, the workflow will be considered to start executing.
// This method can be executed multiple times. Those executed later will be merged with those executed earlier.
// The channel with the same name will be reinitialized.
func WithChannelInput[TInput, TOutput any](name string) Option[TInput, TOutput] {
	return func(d *DAG[TInput, TOutput]) error {
		if !d.channels.exists(name) {
			return &ErrChannelNotExist{name: name}
		}
		d.channels.channelInput = name
		return nil
	}
}

// WithChannelOutput specifies the final data output channel for the entire workflow.
//
// The final execution result should be injected into it. After the data is injected into the channel,
// the execution of the workflow is deemed to have ended. This method can be executed multiple times.
// Those executed later will be merged with those executed earlier.
// The channel with the same name will be reinitialized.
func WithChannelOutput[TInput, TOutput any](name string) Option[TInput, TOutput] {
	return func(d *DAG[TInput, TOutput]) error {
		if !d.channels.exists(name) {
			return &ErrChannelNotExist{name: name}
		}
		d.channels.channelOutput = name
		return nil
	}
}

// WithTransits specifies specific nodes for the entire workflow.
//
// This method can be executed multiple times. Those executed later will be merged with those executed earlier.
// The channel with the same name will be reinitialized.
func WithTransits[TInput, TOutput any](transits ...*Transit) Option[TInput, TOutput] {
	return func(d *DAG[TInput, TOutput]) error {
		lenTransits := len(transits)
		if d.transits == nil {
			d.transits = &Transits{workflowTransits: make([]*Transit, lenTransits)}
		}
		if lenTransits == 0 {
			return nil
		}
		for i, t := range transits {
			d.transits.workflowTransits[i] = t
		}
		return nil
	}
}

// WithLogger specifies the Logger for the entire workflow.
//
// This method can be executed multiple times. The ones executed later will overwrite the ones executed earlier.
func WithLogger[TInput, TOutput any](logger LoggerInterface) Option[TInput, TOutput] {
	return func(d *DAG[TInput, TOutput]) error {
		d.logger = logger
		return nil
	}
}

type TransitOption func(d *Transit)

// NewTransit instantiates the workflow's transit.
func NewTransit(name string, options ...TransitOption) *Transit {
	transit := &Transit{name: name}
	for _, option := range options {
		option(transit)
	}
	return transit
}

func WithInputs(names ...string) TransitOption {
	return func(d *Transit) {
		if d.channelInputs == nil {
			d.channelInputs = []string{}
		}
		d.channelInputs = append(d.channelInputs, names...)
	}
}

func WithOutputs(names ...string) TransitOption {
	return func(d *Transit) {
		if d.channelOutputs == nil {
			d.channelOutputs = []string{}
		}
		d.channelOutputs = append(d.channelOutputs, names...)
	}
}

func WithWorker(worker func(context.Context, ...any) (any, error)) TransitOption {
	return func(d *Transit) {
		d.worker = worker
	}
}

type LoggerOption func(d *Logger)

func NewLogger(options ...LoggerOption) *Logger {
	logger := &Logger{}
	if len(options) == 0 {
		options = []LoggerOption{WithLoggerParams(LoggerParams{
			TimestampFormat: "2006-01-02 15:04:05.000000",
			Caller:          true})}
	}
	for _, option := range options {
		option(logger)
	}
	return logger
}

func WithLoggerParams(params LoggerParams) LoggerOption {
	return func(d *Logger) {
		d.params = params
	}
}

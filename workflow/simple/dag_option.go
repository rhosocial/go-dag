// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package simple

import "context"

// Option defines the option used to instantiate a DAG.
type Option[TInput, TOutput any] func(d *DAG[TInput, TOutput]) error

// NewDAG instantiates a workflow.
//
// The new instance does not come with a logger. If you want to specify a logger, use the NewDAGWithLogger method.
func NewDAG[TInput, TOutput any](options ...Option[TInput, TOutput]) (*DAG[TInput, TOutput], error) {
	dag := &DAG[TInput, TOutput]{}
	for _, option := range options {
		err := option(dag)
		if err != nil {
			return nil, err
		}
	}
	return dag, nil
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
		//if !d.channels.exists(name) {
		//	return &ErrChannelNotExist{name: name}
		//}
		if d.channels == nil {
			d.channels = NewDAGChannels()
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
		//if !d.channels.exists(name) {
		//	return &ErrChannelNotExist{name: name}
		//}
		if d.channels == nil {
			d.channels = NewDAGChannels()
		}
		d.channels.channelOutput = name
		return nil
	}
}

// WithDefaultChannels defines the input and output channel for the workflow.
// Note that this method can be placed before or after the WithChannels(). But if it is placed before it,
// you need to ensure that the subsequent WithChannels() does not mention "input" and "output" again.
func WithDefaultChannels[TInput, TOutput any]() Option[TInput, TOutput] {
	return func(d *DAG[TInput, TOutput]) error {
		if d.channels == nil {
			d.channels = NewDAGChannels()
		}
		d.channels.Add("input", "output")
		d.channels.channelInput = "input"
		d.channels.channelOutput = "output"
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

// TransitOption defines the option used to instantiate a Transit.
type TransitOption func(d *Transit)

// NewTransit instantiates the workflow's transit.
func NewTransit(name string, options ...TransitOption) *Transit {
	transit := &Transit{name: name}
	for _, option := range options {
		option(transit)
	}
	return transit
}

// WithInputs specifies the input channels to be monitored for transit.
// Note that the channel names passed in is consistent with the number and order of parameters
// eventually received by the worker.
// This method can be called multiple times,
// but the results of subsequent calls will be merged and overwritten by previous calls.
func WithInputs(names ...string) TransitOption {
	return func(d *Transit) {
		if d.channelInputs == nil {
			d.channelInputs = []string{}
		}
		d.channelInputs = append(d.channelInputs, names...)
	}
}

// WithOutputs specifies the channel names for transit to be injected into worker execution results.
// Each channel injects the same content, so the order of channel names passed in does not matter.
// This method can be called multiple times,
// but the results of subsequent calls will be merged and overwritten by previous calls.
func WithOutputs(names ...string) TransitOption {
	return func(d *Transit) {
		if d.channelOutputs == nil {
			d.channelOutputs = []string{}
		}
		d.channelOutputs = append(d.channelOutputs, names...)
	}
}

// WithWorker specifies a worker for transit.
// This method can be called multiple times, but only the last one takes effect.
func WithWorker(worker func(context.Context, ...any) (any, error)) TransitOption {
	return func(d *Transit) {
		d.worker = worker
	}
}

// LoggerOption defines the option used to instantiate a Logger.
type LoggerOption func(d *Logger)

// NewLogger instantiates a Logger. Its return value can be used as a parameter to WithLogger().
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

// WithLoggerParams defines the logger parameters.
func WithLoggerParams(params LoggerParams) LoggerOption {
	return func(d *Logger) {
		d.params = params
	}
}

// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package simple

// Option defines the option used to instantiate a Workflow.
type Option[TInput, TOutput any] func(d *Workflow[TInput, TOutput]) error

// NewWorkflow instantiates a workflow.
func NewWorkflow[TInput, TOutput any](options ...Option[TInput, TOutput]) (*Workflow[TInput, TOutput], error) {
	dag := &Workflow[TInput, TOutput]{}
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
	return func(d *Workflow[TInput, TOutput]) error {
		if d.channels == nil {
			d.channels = NewWorkflowChannels()
		}
		return d.channels.add(names...)
	}
}

// WithChannelInput specifies the input channel for the entire workflow.
//
// When the workflow is executed, input data is injected from this channel.
// After the input data is injected into the channel, the workflow will be considered to start executing.
// This method can be executed multiple times. Those executed later will be merged with those executed earlier.
// The channel with the same name will be reinitialized.
func WithChannelInput[TInput, TOutput any](name string) Option[TInput, TOutput] {
	return func(d *Workflow[TInput, TOutput]) error {
		//if !d.channels.exists(name) {
		//	return &ErrChannelNotExist{name: name}
		//}
		if d.channels == nil {
			d.channels = NewWorkflowChannels()
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
	return func(d *Workflow[TInput, TOutput]) error {
		if d.channels == nil {
			d.channels = NewWorkflowChannels()
		}
		d.channels.channelOutput = name
		return nil
	}
}

// WithDefaultChannels defines the input and output channel for the workflow.
// Note that this method can be placed before or after the WithChannels(). But if it is placed before it,
// you need to ensure that the subsequent WithChannels() does not mention "input" and "output" again.
func WithDefaultChannels[TInput, TOutput any]() Option[TInput, TOutput] {
	return func(d *Workflow[TInput, TOutput]) error {
		if d.channels == nil {
			d.channels = NewWorkflowChannels()
		}
		err := d.channels.add("input", "output")
		if err != nil {
			return err
		}
		d.channels.channelInput = "input"
		d.channels.channelOutput = "output"
		return nil
	}
}

// WithTransits specifies specific nodes for the entire workflow.
//
// This method can be executed multiple times. Those executed later will be merged with those executed earlier.
// The channel with the same name will be reinitialized.
func WithTransits[TInput, TOutput any](transits ...TransitInterface) Option[TInput, TOutput] {
	return func(d *Workflow[TInput, TOutput]) error {
		lenTransits := len(transits)
		if lenTransits == 0 {
			return nil
		}
		if d.transits == nil {
			d.transits = &Transits{transits: make([]TransitInterface, 0)}
		}
		d.transits.transits = append(d.transits.transits, transits...)
		return nil
	}
}

// WithLoggers specifies the Logger for the entire workflow.
//
// This method can be executed multiple times. The ones executed later will be merged with the ones executed earlier.
func WithLoggers[TInput, TOutput any](loggers ...LoggerInterface) Option[TInput, TOutput] {
	return func(d *Workflow[TInput, TOutput]) error {
		d.muLoggers.Lock()
		defer d.muLoggers.Unlock()
		if d.loggers == nil {
			d.loggers = &Loggers{}
		}
		for _, logger := range loggers {
			d.loggers.loggers = append(d.loggers.loggers, logger)
		}
		return nil
	}
}

// TransitOption defines the option used to instantiate a Transit.
type TransitOption func(d TransitInterface)

// LoggerOption defines the option used to instantiate a Logger.
type LoggerOption func(d *Logger)

// NewLogger instantiates a Logger. Its return value can be used as a parameter to WithLoggers().
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

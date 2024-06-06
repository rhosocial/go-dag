// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package workflow

import "github.com/rhosocial/go-dag/workflow/standard/context"

// Context defines an interface that extends the standard context.Context interface
// with additional methods specific to managing channels within the workflow.
type Context interface {
	context.Context
	GetChannelInput() string
	GetChannelOutput() string
	BuildChannels()
	AppendChannels(names ...string)
	GetChannel(name string) (chan any, bool)
}

// workflowContext is a concrete implementation of the Context interface
// that provides methods for managing channels within a workflow.
type workflowContext struct {
	context.BaseContext

	// channels holds a map of channel names to channels.
	channels map[string]chan any

	// channelInput defines the channel used for input. Currently, only a single input channel is supported.
	channelInput string
	// channelOutput defines the channel used for output. Currently, only a single output channel is supported.
	channelOutput string
}

// GetChannelInput returns the name of the input channel.
func (w *workflowContext) GetChannelInput() string {
	return w.channelInput
}

// GetChannelOutput returns the name of the output channel.
func (w *workflowContext) GetChannelOutput() string {
	return w.channelOutput
}

// BuildChannels initializes the channels map.
func (w *workflowContext) BuildChannels() {
	w.channels = make(map[string]chan any)
}

// AppendChannels adds new channels to the channels map if they do not already exist.
func (w *workflowContext) AppendChannels(names ...string) {
	for _, name := range names {
		if _, ok := w.channels[name]; !ok {
			w.channels[name] = make(chan any)
		}
	}
}

// GetChannel retrieves a channel by name from the channels map.
// It returns the channel and a boolean indicating whether the channel was found.
func (w *workflowContext) GetChannel(name string) (chan any, bool) {
	if w.channels == nil {
		return nil, false
	}
	if ch, ok := w.channels[name]; ok {
		return ch, true
	}
	return nil, false
}

// NewContext creates a new workflowContext with the specified input and output channels
// and any additional context options. It returns the created workflowContext and any error encountered.
// "input" and "output" are the names of the input and output channels of the workflow.
func NewContext(input string, output string, options ...context.Option) (*workflowContext, error) {
	baseCtx, err := context.NewContext(options...)
	if err != nil {
		return nil, err
	}
	ctx := &workflowContext{
		BaseContext:   *baseCtx,
		channels:      make(map[string]chan any),
		channelInput:  input,
		channelOutput: output,
	}
	return ctx, nil
}

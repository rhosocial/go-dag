// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package logger

// Interface defines the methods that should be implemented to record events.
type Interface interface {
	// Log records the event.
	//
	// Note that this method is actually executed asynchronously.
	Log(event EventInterface)
}

// Logger implements a simple log transmitter.
type Logger struct {
	// eventChannel stores the event manager's event receiving channel.
	//
	// Note that this channel must be initialized, otherwise it will cause a panicking.
	eventChannel chan EventInterface
	Interface
}

// Log records the event.
//
// Note that you must first ensure that the corresponding event collector has enabled listening.
// Otherwise, calling this method will be blocked.
func (l *Logger) Log(event EventInterface) {
	l.eventChannel <- event
}

// NewLogger instantiates a new logger.
//
//   - eventChannel: the destination to which the event is sent.
func NewLogger(eventChannel chan EventInterface) Interface {
	return &Logger{
		eventChannel: eventChannel,
	}
}

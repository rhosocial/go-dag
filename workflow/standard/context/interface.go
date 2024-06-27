// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package context provides a flexible context implementation for managing cancellations,
// deadlines, and request-scoped values across Goroutines.
package context

import (
	"context"

	"github.com/rhosocial/go-dag/workflow/standard/logger"
)

// Context defines the methods that a context implementation must satisfy.
type Context interface {
	// Cancel cancels the context with the provided error cause.
	Cancel(cause error)
	GetContext() context.Context
	GetIdentifier() IdentifierInterface
	GetOptions() OptionsInterface
	GetReports() ReportsInterface
	GetEventManager() logger.EventManagerInterface
}

// BaseContext represents a context instance that encapsulates a context.BaseContext and
// additional functionalities for managing cancellations and request-scoped values.
//
// BaseContext is used to represent the context of a single execution. In addition to
// holding the context and cancel function, it carries the identifier for the execution,
// all configuration options, and the post-execution reports. This BaseContext is used to
// distinguish between different executions fed into the same workflow.
//
// The eventManager field provides functionality for handling events sent by workflow or transit workers
// and dispatching them to registered subscribers. This enables asynchronous communication and event-driven
// processing within the workflow. If eventManager is not specified, events will not be listened to or sent
// to any subscribers for this execution, making it an optional field depending on the needs of the workflow.
type BaseContext struct {
	// context holds the standard context.BaseContext instance.
	// This field is exported to the ctx parameter passed to the worker in the transit.
	context context.Context

	// cancel holds the cancel function that can cancel the context with a specific cause.
	cancel context.CancelCauseFunc

	Context

	// identifier holds an instance of IdentifierInterface for managing unique identifiers.
	identifier IdentifierInterface

	// options holds an instance of OptionsInterface for storing configuration options.
	//
	// The options are used to adjust the behavior of both global and specific transit executions,
	// but do not store parameters passed to the transit's worker. Therefore, the worker in the transit
	// cannot see the current options, meaning these options cannot be used to pass parameters to the worker.
	//
	// This field is optional; if not specified, it means there are no additional parameters for this execution.
	options OptionsInterface

	// reports holds an instance of ReportsInterface for managing the reports of the current execution.
	//
	// This reports instance is used to collect content reported by the workflow during execution,
	// and thus it is not visible to the worker in the transit.
	//
	// This field is optional; if not specified, it means no reports will be collected for this execution.
	reports ReportsInterface

	// eventManager holds an instance of EventManagerInterface for managing events and subscribers.
	//
	// The eventManager is responsible for handling events sent by the workflow or transit workers
	// and dispatching them to the registered subscribers. This allows for asynchronous communication
	// and event-driven processing within the workflow. Subscribers can listen for specific events
	// and act upon them when they occur.
	//
	// This field is optional; if not specified, it means events will not be listened to or sent
	// to any subscribers for this execution. If event-driven communication is not needed,
	// this field can be left unspecified.
	eventManager logger.EventManagerInterface
}

// Cancel cancels the context with the provided error cause.
func (c *BaseContext) Cancel(cause error) {
	c.cancel(cause)
}

func (c *BaseContext) GetContext() context.Context { return c.context }

func (c *BaseContext) GetIdentifier() IdentifierInterface { return c.identifier }

func (c *BaseContext) GetOptions() OptionsInterface { return c.options }

func (c *BaseContext) GetReports() ReportsInterface { return c.reports }

func (c *BaseContext) GetEventManager() logger.EventManagerInterface { return c.eventManager }

// Option is a function type for defining context configuration options.
type Option func(*BaseContext) error

// NewContext creates a new context instance with the given configuration options.
func NewContext(options ...Option) (*BaseContext, error) {
	c := &BaseContext{}
	for _, option := range options {
		err := option(c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

type CancelFuncNilError struct {
	error
}

func (c CancelFuncNilError) Error() string { return "cancel function is nil" }

// WithContext sets the context.BaseContext and cancel function for the context.
func WithContext(context context.Context, cancel context.CancelCauseFunc) Option {
	return func(c *BaseContext) error {
		if cancel == nil {
			return CancelFuncNilError{}
		}
		c.context = context
		c.cancel = cancel
		return nil
	}
}

// WithIdentifier sets the identifier for the context.
func WithIdentifier(identifier IdentifierInterface) Option {
	return func(context *BaseContext) error {
		context.identifier = identifier
		return nil
	}
}

// WithOptions sets the options for the context.
func WithOptions(options OptionsInterface) Option {
	return func(context *BaseContext) error {
		context.options = options
		return nil
	}
}

// WithReports sets the reports for the context.
func WithReports(reports ReportsInterface) Option {
	return func(context *BaseContext) error {
		context.reports = reports
		return nil
	}
}

// WithEventManager sets the event manager for the context.
//
// This option allows you to associate an EventManagerInterface instance with the context,
// enabling event-driven communication and processing within the workflow. The event manager
// is responsible for handling events sent by the workflow or transit workers and dispatching
// them to the registered subscribers.
//
// Note: The context does not automatically start the event manager's listening functionality.
// You must explicitly start the Listen() method on the event manager at the appropriate time
// in your application. Failure to do so will result in the workflow and transit workers being
// blocked when they attempt to send data to the event channel.
func WithEventManager(eventManager logger.EventManagerInterface) Option {
	return func(context *BaseContext) error {
		context.eventManager = eventManager
		return nil
	}
}

// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package context provides a flexible context implementation for managing cancellations,
// deadlines, and request-scoped values across Goroutines.
package context

import "context"

// Interface defines the methods that a context implementation must satisfy.
type Interface interface {
	// Cancel cancels the context with the provided error cause.
	Cancel()
}

// Context represents a context instance that encapsulates a context.Context and
// additional functionalities for managing cancellations and request-scoped values.
type Context struct {
	context context.Context
	cancel  context.CancelCauseFunc

	Interface

	identifier IdentifierInterface
	options    OptionsInterface
}

// Cancel cancels the context with the provided error cause.
func (c *Context) Cancel(cause error) {
	c.cancel(cause)
}

// Option is a function type for defining context configuration options.
type Option func(*Context) error

// NewContext creates a new context instance with the given configuration options.
func NewContext(options ...Option) (*Context, error) {
	c := &Context{}
	for _, option := range options {
		err := option(c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithContext sets the context.Context and cancel function for the context.
func WithContext(context context.Context, cancel context.CancelCauseFunc) Option {
	return func(c *Context) error {
		c.context = context
		c.cancel = cancel
		return nil
	}
}

// WithIdentifier sets the identifier for the context.
func WithIdentifier(identifier IdentifierInterface) Option {
	return func(context *Context) error {
		context.identifier = identifier
		return nil
	}
}

// WithOptions sets the options for the context.
func WithOptions(options OptionsInterface) Option {
	return func(context *Context) error {
		context.options = options
		return nil
	}
}

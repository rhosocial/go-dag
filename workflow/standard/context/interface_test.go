// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package context

import (
	"context"
	"errors"
	"testing"

	"github.com/rhosocial/go-dag/workflow/standard/logger"
	"github.com/stretchr/testify/assert"
)

// Mock implementations for interfaces
type MockIdentifier struct {
	IdentifierInterface
}

func (i MockIdentifier) GetID() string {
	return "mock"
}

func (i MockIdentifier) Equals(other IdentifierInterface) bool {
	return i.GetID() == other.GetID()
}

type MockOptions struct {
	OptionsInterface
}

func (o MockOptions) GetGlobal(key string) (any, error) {
	return nil, nil
}

func (o MockOptions) setGlobal(key string, value any) error {
	return nil
}

func (o MockOptions) GetTransit(transit string, key string) (any, error) {
	return nil, nil
}

func (o MockOptions) setTransit(transit string, key string, value any) error {
	return nil
}

type MockReports struct {
	ReportsInterface
}

func (o MockReports) GetGlobal(key string) (any, error) {
	return nil, nil
}

func (o MockReports) AddGlobal(key string, value any) error {
	return nil
}

func (o MockReports) GetTransit(transit string, key string) (any, error) {
	return nil, nil
}

func (o MockReports) AddTransit(transit string, key string, value any) error {
	return nil
}

type MockEventManager struct {
	logger.EventManagerInterface
}

func (o MockEventManager) Listen(ctx context.Context) {}

func (o MockEventManager) GetLogger() logger.Interface { return nil }

type MockContext struct {
	Context
}

func (m MockContext) Cancel(cause error) {}

func (m MockContext) GetContext() context.Context { return nil }

func (m MockContext) GetIdentifier() IdentifierInterface { return nil }

func (m MockContext) GetOptions() OptionsInterface { return nil }

func (m MockContext) GetReports() ReportsInterface { return nil }

func (m MockContext) GetEventManager() logger.EventManagerInterface { return nil }

// Ensure MockIdentifier implements IdentifierInterface
var _ IdentifierInterface = (*MockIdentifier)(nil)

// Ensure MockOptions implements OptionsInterface
var _ OptionsInterface = (*MockOptions)(nil)

// Ensure MockEventManager implements EventManagerInterface
var _ logger.EventManagerInterface = (*MockEventManager)(nil)

var _ Context = (*MockContext)(nil)

// Mock CancelCauseFunc
func MockCancelCauseFunc(cause error) {}

// Test creating a new BaseContext with different options
func TestNewContext(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())

	// Test with WithContext option
	c1, err := NewContext(WithContext(ctx, cancel))
	assert.NoError(t, err)
	assert.NotNil(t, c1)
	assert.Equal(t, ctx, c1.context)
	assert.NotNil(t, c1.cancel)

	// Test with WithIdentifier option
	mockIdentifier := &MockIdentifier{}
	c2, err := NewContext(WithIdentifier(mockIdentifier))
	assert.NoError(t, err)
	assert.NotNil(t, c2)
	assert.Equal(t, mockIdentifier, c2.identifier)

	// Test with WithOptions option
	mockOptions := &MockOptions{}
	c3, err := NewContext(WithOptions(mockOptions))
	assert.NoError(t, err)
	assert.NotNil(t, c3)
	assert.Equal(t, mockOptions, c3.options)

	// Test with WithOptions option
	mockReports := &MockReports{}
	c4, err := NewContext(WithReports(mockReports))
	assert.NoError(t, err)
	assert.NotNil(t, c4)
	assert.Equal(t, mockReports, c4.reports)

	// Test with WithEventManager option
	mockEventManager := &MockEventManager{}
	c5, err := NewContext(WithEventManager(mockEventManager))
	assert.NoError(t, err)
	assert.NotNil(t, c5)
	assert.Equal(t, mockEventManager, c5.eventManager)

	// Test with multiple options
	c, err := NewContext(
		WithContext(ctx, cancel),
		WithIdentifier(mockIdentifier),
		WithOptions(mockOptions),
		WithReports(mockReports),
		WithEventManager(mockEventManager),
	)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, ctx, c.context)
	assert.Equal(t, ctx, c.GetContext())
	assert.Equal(t, mockIdentifier, c.identifier)
	assert.Equal(t, mockOptions, c.options)
	assert.Equal(t, mockReports, c.reports)
	assert.Equal(t, mockEventManager, c.eventManager)
}

// Test the Cancel method
func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())

	// Wrap cancel function to detect invocation
	cancelInvoked := false
	mockCancel := func(cause error) {
		cancelInvoked = true
		cancel(cause)
	}

	c, err := NewContext(WithContext(ctx, mockCancel))
	assert.NoError(t, err)
	assert.NotNil(t, c)

	// Call Cancel and check if cancel was invoked
	cause := errors.New("test cause")
	c.Cancel(cause)
	assert.True(t, cancelInvoked)
}

func WithContextError(context context.Context, cancel context.CancelCauseFunc) Option {
	return func(c *BaseContext) error {
		return errors.New("test creating new context with error")
	}
}

func TestNewContextWithError(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())

	c, err := NewContext(WithContextError(ctx, cancel))
	assert.Error(t, err)
	assert.Equal(t, "test creating new context with error", err.Error())
	assert.Nil(t, c)
}

func TestNewContextWithNilCancelFunc(t *testing.T) {
	ctx, _ := context.WithCancelCause(context.Background())
	c, err := NewContext(WithContext(ctx, nil))
	assert.ErrorIs(t, err, CancelFuncNilError{})
	assert.Nil(t, c)
	assert.Equal(t, err.Error(), "cancel function is nil")
}

func TestBaseContext_Getter(t *testing.T) {
	ctx, _ := NewContext()
	t.Run("GetContext", func(t *testing.T) {
		ctx.GetContext()
	})
	t.Run("GetIdentifier", func(t *testing.T) {
		ctx.GetIdentifier()
	})
	t.Run("GetOptions", func(t *testing.T) {
		ctx.GetOptions()
	})
	t.Run("GetReports", func(t *testing.T) {
		ctx.GetReports()
	})
	t.Run("GetEventManager", func(t *testing.T) {
		ctx.GetEventManager()
	})
}

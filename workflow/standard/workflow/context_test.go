// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package workflow

import (
	baseContext "context"
	"errors"
	"testing"

	"github.com/rhosocial/go-dag/workflow/standard/context"
	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	baseCtx, _ := context.NewContext(context.WithContext(baseContext.WithCancelCause(baseContext.Background())))
	ctx := workflowContext{
		BaseContext: *baseCtx,
	}
	t.Log(ctx)
}

type ContextError = context.BaseContext

func WithContextError(context baseContext.Context, cancel baseContext.CancelCauseFunc) context.Option {
	return func(c *ContextError) error {
		return errors.New("test creating new context with error")
	}
}

func TestNewContextWithError(t *testing.T) {
	ctx, cancel := baseContext.WithCancelCause(baseContext.Background())
	c, err := NewContext("input", "output", WithContextError(ctx, cancel))
	assert.Error(t, err)
	assert.Equal(t, "test creating new context with error", err.Error())
	assert.Nil(t, c)
}

func TestWorkflowContext_GetChannel(t *testing.T) {
	t.Run("channel not initialized", func(t *testing.T) {
		ctx := workflowContext{}
		ch, exist := ctx.GetChannel("input")
		assert.Nil(t, ch)
		assert.False(t, exist)
	})
	t.Run("channel initialized", func(t *testing.T) {
		ctx := workflowContext{
			channels: make(map[string]chan any),
		}
		ch, exist := ctx.GetChannel("input")
		assert.Nil(t, ch)
		assert.False(t, exist)
	})
}

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
	ctx := Context{
		Context: *baseCtx,
	}
	t.Log(ctx)
}

type ContextError = context.Context

func WithContextError(context baseContext.Context, cancel baseContext.CancelCauseFunc) context.Option {
	return func(c *ContextError) error {
		return errors.New("test creating new context with error")
	}
}

func TestNewContextWithError(t *testing.T) {
	ctx, cancel := baseContext.WithCancelCause(baseContext.Background())
	c, err := NewContext(WithContextError(ctx, cancel))
	assert.Error(t, err)
	assert.Equal(t, "test creating new context with error", err.Error())
	assert.Nil(t, c)
}

func TestChannelNotInitializedError_Error(t *testing.T) {
	err := ChannelNotInitializedError{}
	assert.Equal(t, "channel not initialized", err.Error())
}

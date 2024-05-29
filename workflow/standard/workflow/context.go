// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package workflow

import (
	"sync"

	"github.com/rhosocial/go-dag/workflow/standard/context"
)

type Context struct {
	context.Context

	channels map[string]chan any
	wg       sync.WaitGroup

	// channelInput defines the channel used for input. Currently only a single input channel is supported.
	channelInput string
	// channelOutput defines the channel used for output. Currently only a single output channel is supported.
	channelOutput string
}

func NewContext(options ...context.Option) (*Context, error) {
	baseCtx, err := context.NewContext(options...)
	if err != nil {
		return nil, err
	}
	ctx := &Context{
		Context:  *baseCtx,
		channels: make(map[string]chan any),
	}
	return ctx, nil
}

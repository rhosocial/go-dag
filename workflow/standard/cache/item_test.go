// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockItem struct {
	ItemInterface
}

func (m *MockItem) Expired() bool { return false }

func (m *MockItem) Value() any { return nil }

var _ ItemInterface = (*MockItem)(nil)

func TestNewItem(t *testing.T) {
	item := NewItem(int64(1))
	assert.NotNil(t, item)

	assert.Equal(t, item.Value(), int64(1))
	assert.Equal(t, item.Expired(), false)
}

func TestItem_Expired(t *testing.T) {
	expiredAt := time.Now().Add(-time.Second)
	item := NewItem(int64(1))
	item.expiredAt = &expiredAt

	assert.Equal(t, item.Expired(), true)
	assert.Equal(t, item.Value(), int64(1))

	expiredAt = time.Now().Add(time.Second)
	assert.Equal(t, item.Expired(), false)
}

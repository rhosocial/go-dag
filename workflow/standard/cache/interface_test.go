// Copyright (c) 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package cache

type MockCache struct {
	Interface
}

func (c *MockCache) Get(key KeyGetter) (any, error) { return nil, nil }

func (c *MockCache) Set(key KeyGetter, value any, options ...ItemOption) error { return nil }

func (c *MockCache) Delete(key KeyGetter) error { return nil }

func (c *MockCache) Clear() error { return nil }

var _ Interface = (*MockCache)(nil)

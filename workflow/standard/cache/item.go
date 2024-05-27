// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package cache

import "time"

// ItemInterface defines an interface for cache items.
type ItemInterface interface {
	// Expired checks if the item has expired.
	//
	// Returns
	//
	// - bool: true if the item has expired, otherwise false.
	Expired() bool

	// Value retrieves the value of the item.
	//
	// Returns
	//
	// - any: the value of the item.
	Value() any
}

// Item represents a cache item.
type Item struct {
	value     any
	expiredAt *time.Time
	ItemInterface
}

// Expired checks if the item has expired.
func (i Item) Expired() bool {
	if i.expiredAt == nil {
		return false
	}
	return i.expiredAt.Before(time.Now())
}

// Value retrieves the value of the item.
func (i Item) Value() any {
	return i.value
}

// NewItem creates and initializes a new cache item with the given value.
func NewItem(value any) Item {
	return Item{value: value}
}

// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package cache

import "time"

// ErrInvalidTTL represents an error indicating an invalid Time To Live (TTL) value.
type ErrInvalidTTL struct {
	error
}

// Error returns the error message for ErrInvalidTTL.
func (e ErrInvalidTTL) Error() string {
	return "invalid TTL in cache"
}

// ErrInvalidExpiration represents an error indicating an invalid expiration time.
type ErrInvalidExpiration struct {
	error
}

// Error returns the error message for ErrInvalidExpiration.
func (e ErrInvalidExpiration) Error() string {
	return "invalid expiration in cache"
}

// ItemOption represents an option function that can be applied to cache items.
type ItemOption func(*Item) error

// WithTTL creates an ItemOption with a specified duration as the Time To Live (TTL) for the cache item.
func WithTTL(duration time.Duration) ItemOption {
	return func(i *Item) error {
		if duration <= 0 {
			return ErrInvalidTTL{}
		}
		expiredAt := time.Now().Add(duration)
		i.expiredAt = &expiredAt
		return nil
	}
}

// WithExpiredAt creates an ItemOption with a specified expiration time for the cache item.
func WithExpiredAt(expiredAt time.Time) ItemOption {
	return func(i *Item) error {
		if expiredAt.Before(time.Now()) {
			return ErrInvalidExpiration{}
		}
		i.expiredAt = &expiredAt
		return nil
	}
}

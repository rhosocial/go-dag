// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package context

import (
	"fmt"
	"math/rand"
	"time"
)

// IdentifierInterface interface, requires implementation of GetID and Equals methods
type IdentifierInterface interface {
	GetID() string
	Equals(other IdentifierInterface) bool
}

// IntID implementation
type IntID struct {
	ID int64
}

func (i IntID) GetID() string {
	return fmt.Sprintf("%d", i.ID)
}

func (i IntID) Equals(other IdentifierInterface) bool {
	return i.GetID() == other.GetID()
}

// StringID implementation
type StringID struct {
	ID string
}

func (s StringID) GetID() string {
	return s.ID
}

func (s StringID) Equals(other IdentifierInterface) bool {
	return s.GetID() == other.GetID()
}

// NewIdentifier factory method
func NewIdentifier(value any) IdentifierInterface {
	switch v := value.(type) {
	case int:
		if v > 0 {
			return IntID{ID: int64(v)}
		}
	case int64:
		if v > 0 {
			return IntID{ID: v}
		}
	case string:
		return StringID{ID: v}
	}
	return IntID{ID: generateRandomInt64()}
}

// generateRandomInt64 generates a random int64 number
func generateRandomInt64() int64 {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand.Int63()
}

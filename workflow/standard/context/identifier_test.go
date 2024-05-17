// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package context

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewIdentifier tests the NewIdentifier function
func TestNewIdentifier(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{123, "IntID"},
		{int64(456), "IntID"},
		{int64(-123), "IntID"},
		{0, "IntID"},
		{nil, "IntID"},
		{"string", "StringID"},
	}

	for _, test := range tests {
		id := NewIdentifier(test.input)
		switch v := id.(type) {
		case IntID:
			if test.expected != "IntID" {
				t.Errorf("Expected %s but got IntID", test.expected)
			}
			if test.input == nil || (isNegativeInt(test.input)) || test.input == 0 || isString(test.input) {
				if !isInt64(v.GetID()) {
					t.Errorf("Expected a random int64 but got %s", v.GetID())
				}
			} else {
				expected := fmt.Sprintf("%d", test.input)
				if v.GetID() != expected {
					t.Errorf("Expected ID %s but got %s", expected, v.GetID())
				}
			}
		case StringID:
			if test.expected != "StringID" {
				t.Errorf("Expected %s but got StringID", test.expected)
			}
			assert.Equal(t, test.input, id.GetID())
		default:
			t.Errorf("Unexpected type %T", v)
		}
	}
}

func TestIdentifierEquals(t *testing.T) {
	tests := []struct {
		target any
		expect any
		equals bool
	}{
		{"123", 123, true},
		{123, "123", true},
		{"123", "123", true},
		{123, 123, true},
		{"123", "1234", false},
		{123, 1234, false},
		{"123", 1234, false},
		{123, "1234", false},
	}
	for _, test := range tests {
		target := NewIdentifier(test.target)
		expect := NewIdentifier(test.expect)
		assert.Equal(t, test.equals, target.Equals(expect), target.GetID(), expect.GetID(), test.equals)
	}
}

// isInt64 checks if the string is a valid int64
func isInt64(str string) bool {
	_, err := strconv.ParseInt(str, 10, 64)
	return err == nil
}

// isNegativeInt checks if the value is a negative integer
func isNegativeInt(value any) bool {
	switch v := value.(type) {
	case int:
		return v < 0
	case int64:
		return v < 0
	}
	return false
}

// isString checks if the value is a string
func isString(value any) bool {
	_, ok := value.(string)
	return ok
}

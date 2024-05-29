// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCancelError(t *testing.T) {
	err := CancelError{
		Message: "workflow canceled",
	}
	assert.Equal(t, "context cancelled: workflow canceled", err.Error())
}

// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package context

import "fmt"

// CancelError represents an error that causes the context to be cancelled.
type CancelError struct {
	Message string
}

func (e *CancelError) Error() string {
	return fmt.Sprintf("context cancelled: %s", e.Message)
}

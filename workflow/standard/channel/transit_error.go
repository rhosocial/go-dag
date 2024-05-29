// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package channel

import "fmt"

type WorkerNotSpecifiedError struct {
	name string
	error
}

func NewWorkerNotSpecifiedError(name string) *WorkerNotSpecifiedError {
	return &WorkerNotSpecifiedError{
		name: name,
	}
}

func (e *WorkerNotSpecifiedError) Error() string {
	return fmt.Sprintf("worker %s not specified", e.name)
}

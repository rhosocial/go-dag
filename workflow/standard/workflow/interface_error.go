// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package workflow

import "fmt"

type GraphNilError struct {
	error
}

func (err GraphNilError) Error() string {
	return "graph is nil"
}

type GraphNotSpecifiedError struct {
	error
}

func (err GraphNotSpecifiedError) Error() string {
	return "graph is not specified"
}

type ChannelNotFoundError struct {
	name string
	error
}

func (err ChannelNotFoundError) Error() string {
	return fmt.Sprintf("channel %s not found", err.name)
}

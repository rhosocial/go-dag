// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package channel

import "fmt"

// HangingIncomingError represents an error where a node has a hanging incoming node.
type HangingIncomingError struct {
	name        string
	predecessor string
	error
}

func (e HangingIncomingError) Error() string {
	return fmt.Sprintf("node %s has a hanging predecessor %s", e.name, e.predecessor)
}

// HangingOutgoingError represents an error where a node has a hanging outgoing node.
type HangingOutgoingError struct {
	name      string
	successor string
	error
}

func (e HangingOutgoingError) Error() string {
	return fmt.Sprintf("node %s has a hanging successor %s", e.name, e.successor)
}

// DanglingIncomingError represents an error where a node has no defined incoming node(s).
type DanglingIncomingError struct {
	name string
	error
}

func (e DanglingIncomingError) Error() string {
	return fmt.Sprintf("node %s has no defined predecessor(s)", e.name)
}

// DanglingOutgoingError represents an error where a node has no defined outgoing node(s).
type DanglingOutgoingError struct {
	name string
	error
}

func (e DanglingOutgoingError) Error() string {
	return fmt.Sprintf("node %s has no defined successor(s)", e.name)
}

// SourceDuplicatedError represents an error where a source node is duplicated.
type SourceDuplicatedError struct {
	name   string
	source string
	error
}

func (e SourceDuplicatedError) Error() string {
	return fmt.Sprintf("node %s cannot be source as the source %s already exists", e.name, e.source)
}

// SinkDuplicatedError represents an error where a sink node is duplicated.
type SinkDuplicatedError struct {
	name string
	sink string
	error
}

func (e SinkDuplicatedError) Error() string {
	return fmt.Sprintf("node %s cannot be sink as the sink %s already exists", e.name, e.sink)
}

// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package workflow

type ChannelNotInitializedError struct {
	error
}

func (err ChannelNotInitializedError) Error() string {
	return "channel not initialized"
}

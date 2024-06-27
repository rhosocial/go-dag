// Copyright (c) 2023 - 2024. vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package worker

import "github.com/rhosocial/go-dag/workflow/standard/logger"

type EventRenamingWorker struct {
	oldID string
	newID string
	logger.EventInterface
}

func NewEventRenamingWorker(oldID, newID string) *EventRenamingWorker {
	return &EventRenamingWorker{
		oldID: oldID,
		newID: newID,
	}
}

type EventShutdown struct {
	logger.EventInterface
}

func NewEventShutdown() *EventShutdown {
	return &EventShutdown{}
}

type EventStartingWorker struct {
	ID string
	logger.EventInterface
}

func NewEventStartingWorker(id string) *EventStartingWorker {
	return &EventStartingWorker{
		ID: id,
	}
}

type EventResizing struct {
	oldSize int
	newSize int
	logger.EventInterface
}

func NewEventResizing(oldSize, newSize int) *EventResizing {
	return &EventResizing{
		oldSize: oldSize,
		newSize: newSize,
	}
}

type EventStoppingWorker struct {
	ID string
	logger.EventInterface
}

func NewEventStoppingWorker(id string) *EventStoppingWorker {
	return &EventStoppingWorker{
		ID: id,
	}
}

type EventExitingWorker struct {
	ID         string
	StopWorker bool
	logger.EventInterface
}

func NewEventExitingWorker(id string, stopWorker bool) *EventExitingWorker {
	return &EventExitingWorker{
		ID:         id,
		StopWorker: stopWorker,
	}
}

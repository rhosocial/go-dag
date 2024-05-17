// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package context

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockEvent represents a mock event for testing purposes.
type MockEvent struct {
	Message string
}

// MockSubscriber represents a mock subscriber for testing purposes.
type MockSubscriber struct {
	ReceivedEvents []EventInterface
}

// ReceiveEvent implements the ReceiveEvent method of SubscriberInterface.
func (s *MockSubscriber) ReceiveEvent(event EventInterface) {
	s.ReceivedEvents = append(s.ReceivedEvents, event)
}

func TestEventManager(t *testing.T) {
	// Create mock subscriber
	subscriber := &MockSubscriber{}

	// Create a new eventManager with a subscriber
	em, err := NewEventManager(WithSubscriber("mock_subscriber", subscriber))
	if err != nil {
		t.Fatalf("Error creating eventManager: %v", err)
	}

	ctxBg, cancelBg := context.WithCancel(context.Background())
	go em.Listen(ctxBg)

	// Create a context with eventManager
	ctx := WorkerContextWithEventManager(context.Background(), em)

	// Retrieve event channel from context
	eventChannel := GetEventChannelFromWorkerContext(ctx)
	if eventChannel == nil {
		t.Fatal("Failed to retrieve event channel from context")
	}

	// Publish an event
	event := &MockEvent{Message: "Test event"}
	eventChannel <- event

	// Wait for a short time to ensure event processing
	time.Sleep(100 * time.Millisecond)

	// Check if the subscriber received the event
	if len(subscriber.ReceivedEvents) != 1 {
		t.Fatalf("Expected 1 received event, got %d", len(subscriber.ReceivedEvents))
	}

	// Check the content of the received event
	receivedEvent := subscriber.ReceivedEvents[0].(*MockEvent)
	expectedMessage := "Test event"
	if receivedEvent.Message != expectedMessage {
		t.Fatalf("Expected event message '%s', got '%s'", expectedMessage, receivedEvent.Message)
	}

	// Remove subscriber from eventManager
	delete(em.subscribers, "mock_subscriber")

	// Publish another event after removing subscriber
	eventChannel <- &MockEvent{Message: "Test event 2"}

	// Wait for a short time to ensure event processing
	time.Sleep(100 * time.Millisecond)

	// Ensure that subscriber did not receive the second event
	if len(subscriber.ReceivedEvents) != 1 {
		t.Fatalf("Expected 1 received event, got %d", len(subscriber.ReceivedEvents))
	}

	cancelBg()
}

func TestContextWithEventManager(t *testing.T) {
	// Create a new eventManager
	em, err := NewEventManager()
	if err != nil {
		t.Fatalf("Error creating eventManager: %v", err)
	}

	// Create a context with eventManager
	ctx := WorkerContextWithEventManager(context.Background(), em)

	// Retrieve event channel from context
	retrieveChan := GetEventChannelFromWorkerContext(ctx)
	if retrieveChan == nil {
		t.Fatal("Failed to retrieve eventManager from context")
	}

	// Ensure that retrieved eventManager matches the original one
	if retrieveChan != em.eventChannel {
		t.Fatal("Retrieved eventManager does not match the original one")
	}
}

func WithSubscriberError() EventManagerOption {
	return func(em *EventManager) error {
		return errors.New("with subscriber error")
	}
}

func TestEventManagerWithError(t *testing.T) {
	em, err := NewEventManager(WithSubscriberError())
	assert.Error(t, err)
	assert.Nil(t, em)
}

func TestGetEventChannelFromContextNil(t *testing.T) {
	// Retrieve event channel from context
	retrieveChan := GetEventChannelFromWorkerContext(context.Background())
	assert.Nil(t, retrieveChan)
}

// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

package context

import "context"

// EventInterface represents an event sent by workflow or transit workers.
type EventInterface interface{}

// SubscriberInterface represents a subscriber that listens for events.
type SubscriberInterface interface {
	// ReceiveEvent is called when an event is received.
	ReceiveEvent(event EventInterface)
}

// EventManagerInterface represents an interface for managing events.
type EventManagerInterface interface {
	// Listen listens for events and sends them to subscribers.
	Listen(ctx context.Context)
}

// EventManager manages events and subscribers.
type EventManager struct {
	// eventChannel is a channel through which events are sent.
	eventChannel chan EventInterface

	// subscribers is a map of subscriber identifiers to subscriber instances.
	subscribers map[string]SubscriberInterface
}

// NewEventManager creates a new EventManager instance.
func NewEventManager(options ...EventManagerOption) (*EventManager, error) {
	eventManager := &EventManager{
		// Initialize eventChannel to handle EventInterface.
		eventChannel: make(chan EventInterface),

		// Initialize subscribers as an empty map.
		subscribers: make(map[string]SubscriberInterface),
	}
	for _, option := range options {
		// Apply each option to the event manager.
		err := option(eventManager)
		if err != nil {
			return nil, err
		}
	}
	return eventManager, nil
}

// EventManagerOption defines a type for configuring EventManager.
type EventManagerOption func(*EventManager) error

// WithSubscriber adds a subscriber to the event manager.
func WithSubscriber(identifier string, subscriber SubscriberInterface) EventManagerOption {
	return func(manager *EventManager) error {
		manager.subscribers[identifier] = subscriber
		return nil
	}
}

// Listen starts a goroutine to receive events sent by workflow and transit workers,
// and dispatches them to all subscribers. This method needs to be asynchronously started
// to avoid blocking, and should be initiated before executing the workflow. The parameter ctx
// controls the context in which this goroutine stops, and it is not the same context as the one
// used for executing the workflow. For instance, after one execution of the workflow finishes,
// its context is destroyed, but you may need to retain the listening goroutine for a while longer
// to ensure that all events are successfully dispatched. On the other hand, the event manager
// can also be reused, and its context is related to the lifecycle of the event manager.
func (em *EventManager) Listen(ctx context.Context) {
	for {
		select {
		case event := <-em.eventChannel:
			// Dispatch event to all subscribers.
			for _, subscriber := range em.subscribers {
				if subscriber != nil {
					// Call ReceiveEvent in a separate goroutine.
					go subscriber.ReceiveEvent(event)
				}
			}
		case <-ctx.Done():
			// Exit the loop if the context is done.
			return
		default:
		}
	}
}

// KeyEventManagerChannel is the context key for the event manager channel.
const KeyEventManagerChannel = "__go_dag_workflow_event_channel"

// WorkerContextWithEventManager creates a new context with the provided event manager.
func WorkerContextWithEventManager(ctx context.Context, em *EventManager) context.Context {
	return context.WithValue(ctx, KeyEventManagerChannel, em.eventChannel)
}

// GetEventChannelFromWorkerContext retrieves the event channel from the context.
func GetEventChannelFromWorkerContext(ctx context.Context) chan<- EventInterface {
	if em, ok := ctx.Value(KeyEventManagerChannel).(chan EventInterface); ok {
		return em
	}
	return nil
}

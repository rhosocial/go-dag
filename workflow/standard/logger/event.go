// Copyright (c) 2023 - 2024 vistart. All rights reserved.
// Use of this source code is governed by Apache-2.0 license
// that can be found in the LICENSE file.

// Package logger provides an event subscriber and a basic log event.
package logger

import (
	"context"
	"sync"
)

// EventInterface represents an event sent by workflow or transit workers.
//
// If you want to customize events, please extend this interface.
type EventInterface interface{}

// SubscriberInterface represents a subscriber that listens for events.
type SubscriberInterface interface {
	// ReceiveEvent is called when an event is received.
	//
	// Note that if the event is nil, it will not be dispatched.
	ReceiveEvent(event EventInterface)
}

// EventManagerInterface represents an interface for managing events.
type EventManagerInterface interface {
	// Listen listens for events and sends them to subscribers.
	//
	// This method is actually called asynchronously by the event manager,
	// so you don't have to worry about this method being time-consuming and blocking other processes.
	Listen()
	GetLogger() Interface
}

// EventManagerSubscriberInterface represents an interface for managing subscribers.
type EventManagerSubscriberInterface interface {
	AddSubscriber(identifier string, subscriber SubscriberInterface)
	RemoveSubscriber(identifier string)
	GetSubscriber(identifier string) SubscriberInterface
}

// EventManager manages events and subscribers.
type EventManager struct {
	// eventChannel is a channel through which events are sent.
	eventChannel chan EventInterface

	// subscribers is a map of subscriber identifiers to subscriber instances.
	subscribers map[string]SubscriberInterface
	mu          sync.RWMutex

	// logger holds the log transport that sends logs to this event manager.
	// Please see GetLogger
	logger Interface

	// ctx is context for Listen.
	ctx context.Context
}

// NewEventManager creates a new EventManager instance.
//
// Note that the new instance cannot receive events directly and needs to start listening first. For example:
//
//	ctxBg, cancelBg := context.WithCancel(context.Background())
//	manager, err := NewEventManager(WithSubscriber("identifier", subscriber), WithListeningContext(ctxBg))
//	go manager.Listen()
//	<receive events...>
//
// You can also not pass in a context, and use context.Background by default,
// which means that event collection will not stop until the process terminates.
func NewEventManager(options ...EventManagerOption) (*EventManager, error) {
	eventManager := &EventManager{
		// Initialize eventChannel to handle EventInterface.
		eventChannel: make(chan EventInterface),

		// Initialize subscribers as an empty map.
		subscribers: make(map[string]SubscriberInterface),
	}
	eventManager.logger = NewLogger(eventManager.eventChannel)
	for _, option := range options {
		// Apply each option to the event manager.
		err := option(eventManager)
		if err != nil {
			return nil, err
		}
	}
	if eventManager.ctx == nil {
		eventManager.ctx = context.Background()
	}
	return eventManager, nil
}

// EventManagerOption defines a type for configuring EventManager.
type EventManagerOption func(*EventManager) error

// WithSubscriber adds a subscriber to the event manager.
//
//   - identifier: represents the identifier of the subscriber.
//   - subscriber: the subscriber instance, which needs to implement the SubscriberInterface.
//
// This method can be called multiple times to add multiple subscribers. When an event is received,
// it will be dispatched to each subscriber.
// The dispatch order is not guaranteed to match the order in which subscribers were added,
// so you cannot rely on the order of subscriber addition to determine the event dispatch order.
// If no subscribers are added, received events will be discarded and not dispatched.
func WithSubscriber(identifier string, subscriber SubscriberInterface) EventManagerOption {
	return func(manager *EventManager) error {
		manager.AddSubscriber(identifier, subscriber)
		return nil
	}
}

// WithListeningContext set the context for the `Listen()` method.
func WithListeningContext(ctx context.Context) EventManagerOption {
	return func(manager *EventManager) error {
		manager.ctx = ctx
		return nil
	}
}

// Listen starts a goroutine to receive events sent by workflow and transit workers,
// and dispatches them to all subscribers. This method needs to be asynchronously started to avoid blocking,
// and should be initiated before executing the workflow. The parameter ctx of the event manager
// controls the context in which this goroutine stops, and it is not the same context as the one
// used for executing the workflow. For instance, after one execution of the workflow finishes,
// its context is destroyed, but you may need to retain the listening goroutine for a while longer
// to ensure that all events are successfully dispatched. On the other hand, the event manager
// can also be reused, and its context is related to the lifecycle of the event manager.
func (em *EventManager) Listen() {
	for {
		select {
		case event := <-em.eventChannel:
			// Dispatch event to all subscribers.
			em.mu.Lock()
			for _, subscriber := range em.subscribers {
				if subscriber != nil {
					// Call ReceiveEvent in a separate goroutine.
					go subscriber.ReceiveEvent(event)
				}
			}
			em.mu.Unlock()
		case <-em.ctx.Done():
			// Exit the loop if the context is done.
			return
		default:
		}
	}
}

// GetLogger returns the logger instance.
func (em *EventManager) GetLogger() Interface {
	return em.logger
}

// AddSubscriber adds a subscriber with specified identifier.
//
// If the specified subscriber exists, it is replaced.
func (em *EventManager) AddSubscriber(identifier string, subscriber SubscriberInterface) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.subscribers[identifier] = subscriber
}

// RemoveSubscriber removes a subscriber with specified identifier.
//
// If the specified subscriber does not exist, nothing happens.
func (em *EventManager) RemoveSubscriber(identifier string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	delete(em.subscribers, identifier)
}

// GetSubscriber gets a subscriber with specified identifier.
//
// If the specified subscriber does not exist, nil will be given.
func (em *EventManager) GetSubscriber(identifier string) SubscriberInterface {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.subscribers[identifier]
}

// KeyEventManagerLogger is ßthe key of the event manager logger in the context.
// @experimental
const KeyEventManagerLogger = "__go_dag_workflow_event_logger"

// WorkerContextWithLogger derives a context with value from the specified context and stores the logger in it.
// @experimental
func WorkerContextWithLogger(ctx context.Context, logger Interface) context.Context {
	return context.WithValue(ctx, KeyEventManagerLogger, logger)
}

// GetLoggerFromWorkerContext gets a logger from the specified context.
// @experimental
func GetLoggerFromWorkerContext(ctx context.Context) Interface {
	if logger, ok := ctx.Value(KeyEventManagerLogger).(Interface); ok {
		return logger
	}
	return nil
}

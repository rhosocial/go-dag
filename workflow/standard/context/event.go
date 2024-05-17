package context

import "context"

// EventInterface represents an event sent by workflow or transit workers.
type EventInterface interface{}

// SubscriberInterface represents a subscriber that listens for events.
type SubscriberInterface interface {
	// ReceiveEvent is called when an event is received.
	ReceiveEvent(event EventInterface)
}

type EventManagerInterface interface {
	Listen(ctx context.Context)
}

// EventManager manages events and subscribers.
type EventManager struct {
	eventChannel chan EventInterface
	subscribers  map[string]SubscriberInterface
}

// NewEventManager creates a new EventManager instance.
func NewEventManager(options ...EventManagerOption) (*EventManager, error) {
	eventManager := &EventManager{
		eventChannel: make(chan EventInterface),
		subscribers:  make(map[string]SubscriberInterface),
	}
	for _, option := range options {
		err := option(eventManager)
		if err != nil {
			return nil, err
		}
	}
	return eventManager, nil
}

type EventManagerOption func(*EventManager) error

// WithSubscriber adds a subscriber to the event manager.
func WithSubscriber(identifier string, subscriber SubscriberInterface) EventManagerOption {
	return func(manager *EventManager) error {
		manager.subscribers[identifier] = subscriber
		return nil
	}
}

// Listen sends an event to all subscribers.
func (em *EventManager) Listen(ctx context.Context) {
	for {
		select {
		case event := <-em.eventChannel:
			for _, subscriber := range em.subscribers {
				if subscriber != nil {
					go subscriber.ReceiveEvent(event)
				}
			}
		case <-ctx.Done():
			return
		default:
		}
	}
}

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

package utils

import (
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"sync"
)

const (
	// ErrorSubscriber is return when an error occurs while processing an event in a subscriber.
	ErrorSubscriber = "subscriber error"
)

// Observer describes the observer pattern.
type Observer interface {
	Publish(eventName string, eventData interface{}) error
	Subscribe(eventName string, eventSubscriber EventSubscriber)
}

// EventSubscriber is a callback function for observed events.
type EventSubscriber func(eventName string, eventData interface{}) error

type syncObserver struct {
	mutex       *sync.Mutex
	subscribers map[string][]EventSubscriber
}

// NewSyncObserver instantiates a new synchronous, thread-safe observer.
func NewSyncObserver() Observer {
	return &syncObserver{
		mutex:       &sync.Mutex{},
		subscribers: make(map[string][]EventSubscriber),
	}
}

// Publish implements Observer interface.
func (o *syncObserver) Publish(eventName string, eventData interface{}) error {
	subscribers := o.safeGetSubscribers(eventName)
	if subscribers == nil {
		return nil
	}
	for _, subscriber := range subscribers {
		if err := subscriber(eventName, eventData); err != nil {
			return xerror.Wrap(err, ErrorSubscriber)
		}
	}
	return nil
}

func (o *syncObserver) safeGetSubscribers(eventName string) []EventSubscriber {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if subscribers, ok := o.subscribers[eventName]; ok && len(subscribers) > 0 {
		return append(make([]EventSubscriber, 0, len(subscribers)), subscribers...)
	}
	return nil
}

// Subscribe implement the Observer interface.
func (o *syncObserver) Subscribe(eventName string, eventSubscriber EventSubscriber) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.subscribers[eventName] = append(o.subscribers[eventName], eventSubscriber)
}

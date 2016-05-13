package observer

import "sync"

//Observer interface
type Observer interface {
	Publish(event string, data interface{}) error
	Subscribe(event string, eventListener EventListener)
}

//EventListener is callback function
type EventListener func(data interface{}) error

//ObserverImpl implements Observer interface
type ObserverImpl struct {
	mutex     sync.Mutex
	listeners map[string][]EventListener
}

//New instantiates ObserverImpl
func New() *ObserverImpl {
	return &ObserverImpl{
		listeners: make(map[string][]EventListener, 0),
	}
}

//Publish implements Observer interface
func (o *ObserverImpl) Publish(event string, data interface{}) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	for _, listener := range o.listeners[event] {
		if err := listener(data); err != nil {
			return err
		}
	}

	return nil
}

//Subscribe implements Observer interface
func (o *ObserverImpl) Subscribe(event string, eventListener EventListener) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.listeners[event] = append(o.listeners[event], eventListener)
}

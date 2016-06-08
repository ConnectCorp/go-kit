package utils

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"testing"
)

func TestSyncObserver(t *testing.T) {
	type recordedEvent struct {
		eventName string
		eventData interface{}
	}

	observer := NewSyncObserver()
	var recordedEvents []*recordedEvent

	eventRecorder := func(eventName string, eventData interface{}) error {
		recordedEvents = append(recordedEvents, &recordedEvent{eventName, eventData})
		return nil
	}

	observer.Subscribe("firstEvent", eventRecorder)
	observer.Subscribe("secondEvent", eventRecorder)

	assert.Nil(t, observer.Publish("firstEvent", "11"))
	assert.Nil(t, observer.Publish("secondEvent", "21"))
	assert.Nil(t, observer.Publish("firstEvent", "12"))
	assert.Nil(t, observer.Publish("secondEvent", "22"))

	assert.Equal(t,
		[]*recordedEvent{
			&recordedEvent{"firstEvent", "11"},
			&recordedEvent{"secondEvent", "21"},
			&recordedEvent{"firstEvent", "12"},
			&recordedEvent{"secondEvent", "22"},
		}, recordedEvents)
}

func TestSyncObserver_SubscriberError(t *testing.T) {
	observer := NewSyncObserver()
	observer.Subscribe("event", func(eventName string, eventData interface{}) error { return xerror.New("err") })
	assert.EqualValues(t, "subscriber error: err", observer.Publish("event", "data").Error())
}

func TestSyncObserver_NoSubscribers(t *testing.T) {
	observer := NewSyncObserver()
	assert.Nil(t, observer.Publish("event", "data"))
}

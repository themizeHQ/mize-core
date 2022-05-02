package events

import (
	"github.com/jiyeyuran/go-eventemitter"
)

var emitter eventemitter.IEventEmitter

func EmitterService() {
	emitter = eventemitter.NewEventEmitter()
	registerEvents(emitter)
}

func registerEvents(e eventemitter.IEventEmitter) {
	for _, event := range EventsArray {
		e.On(event.Name, event.Action)
	}
}

func Emit(event string, args ...interface{}) {
	emitter.Emit(event, args...)
}

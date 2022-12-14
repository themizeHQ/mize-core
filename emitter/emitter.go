package emitter

import (
	"github.com/CHH/eventemitter"
	"go.uber.org/zap"
	"mize.app/logger"
)

type emitter struct {
	e *eventemitter.EventEmitter
}

func (em emitter) Initialise() *emitter {
	emitter := emitter{
		e: eventemitter.New(),
	}
	return &emitter
}

func (em *emitter) Listen(event string, cb interface{}) {
	logger.Info("emitter listening", zap.String("event", event))
	em.e.On(event, cb)
}

func (em *emitter) Emit(event string, data interface{}) {
	logger.Info("emitted event", zap.String("event", event))
	<-em.e.Emit(event, data)
}

var Emitter = emitter.Initialise(emitter{})

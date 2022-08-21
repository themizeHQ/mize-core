package emitter

import (
	"fmt"

	"github.com/CHH/eventemitter"
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
	fmt.Println("listening to...")
	fmt.Println(event)
	em.e.On(event, cb)
}

func (em *emitter) Emit(event string, data interface{}) {
	fmt.Println("emitted")
	<-em.e.Emit(event, data)
}

var Emitter = emitter.Initialise(emitter{})

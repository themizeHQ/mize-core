package events

type event struct {
	Name   string
	Action interface{}
}

var EventsArray = []event{}

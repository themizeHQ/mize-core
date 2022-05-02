package events

type event struct {
	Name   string
	Action interface{}
}

var UserCreatedEvent = event{Name: "USER_CREATED_EVENT", Action: func() {

}}

var EventsArray = []event{UserCreatedEvent}

package eventsqueue

type Event struct {
	Topic   EventTopic
	Payload interface{}
}

type EventTopic string

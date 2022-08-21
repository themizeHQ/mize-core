package emitter

import "fmt"

func EmitterListener() {
	// users
	Emitter.Listen(Events.USER.USER_CREATED, HandleUserCreated)
	Emitter.Listen(Events.USER.USER_VERIFIED, HandleUserVerified)

	// messages

	Emitter.Listen(Events.MESSAGES_EVENTS.MESSAGE_SENT, HandleMessageSent)
	Emitter.Listen(Events.MESSAGES_EVENTS.MESSAGE_DELETED, HandleMessageDeleted)
}

// users
func HandleUserCreated()  {}
func HandleUserVerified() {}

// messages
func HandleMessageSent(data interface{}) {
	fmt.Println(data)
}
func HandleMessageDeleted() {}

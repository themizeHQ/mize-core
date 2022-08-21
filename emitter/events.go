package emitter

type events struct {
	MESSAGES_EVENTS struct {
		MESSAGE_SENT    string
		MESSAGE_DELETED string
	}
	USER struct {
		USER_CREATED  string
		USER_VERIFIED string
	}
}

var Events = events{
	MESSAGES_EVENTS: struct {
		MESSAGE_SENT    string
		MESSAGE_DELETED string
	}{
		MESSAGE_SENT:    "MESSAGE_SENT",
		MESSAGE_DELETED: "MESSAGES_DELETED",
	},
}

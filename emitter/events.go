package emitter

type events struct {
	MESSAGES_EVENTS messageEvent
	AUTH_EVENTS     authEvent
}

type messageEvent struct {
	MESSAGE_SENT    string
	MESSAGE_DELETED string
}
type authEvent struct {
	USER_CREATED  string
	USER_VERIFIED string
	RESEND_OTP    string
}

var Events = events{
	MESSAGES_EVENTS: messageEvent{
		MESSAGE_SENT:    "MESSAGE_SENT",
		MESSAGE_DELETED: "MESSAGES_DELETED",
	},
	AUTH_EVENTS: authEvent{
		USER_CREATED:  "USER_CREATED",
		USER_VERIFIED: "USER_VERIFIED",
		RESEND_OTP:    "RESEND_OTP",
	},
}

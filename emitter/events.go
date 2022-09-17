package emitter

type events struct {
	MESSAGES_EVENTS     messageEvents
	AUTH_EVENTS         authEvents
	NOTIFICATION_EVENTS notificationEvents
}

type messageEvents struct {
	MESSAGE_SENT    string
	MESSAGE_DELETED string
}

type notificationEvents struct {
	NOTIFICATION_CREATED string
	NOTIFICATION_DELETED string
}

type authEvents struct {
	USER_CREATED  string
	USER_VERIFIED string
	RESEND_OTP    string
}

var Events = events{
	MESSAGES_EVENTS: messageEvents{
		MESSAGE_SENT:    "MESSAGE_SENT",
		MESSAGE_DELETED: "MESSAGES_DELETED",
	},
	AUTH_EVENTS: authEvents{
		USER_CREATED:  "USER_CREATED",
		USER_VERIFIED: "USER_VERIFIED",
		RESEND_OTP:    "RESEND_OTP",
	},
	NOTIFICATION_EVENTS: notificationEvents{
		NOTIFICATION_CREATED: "NOTIFICATION_CREATED",
		NOTIFICATION_DELETED: "NOTIFICATION_DELETED",
	},
}

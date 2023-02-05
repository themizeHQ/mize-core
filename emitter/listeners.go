package emitter

import (
	eventsqueue "mize.app/events_queue"
	"mize.app/logger"
)

func EmitterListener() {
	// auth
	Emitter.Listen(Events.AUTH_EVENTS.USER_CREATED, HandleUserCreated)
	Emitter.Listen(Events.AUTH_EVENTS.USER_VERIFIED, HandleUserVerified)
	Emitter.Listen(Events.AUTH_EVENTS.RESEND_OTP, HandleResendOtp)

	// messages
	Emitter.Listen(Events.MESSAGES_EVENTS.MESSAGE_DELETED, HandleMessageDeleted)
	Emitter.Listen(Events.MESSAGES_EVENTS.MESSAGE_SENT, HandleNotifyTaggedUsers)

	// sms
	Emitter.Listen(Events.SMS_EVENTS.SMS_SENT, HandleSMSSent)

	// emails
	Emitter.Listen(Events.EMAIL_EVENTS.EMAIL_SENT, HandleEmailSent)

	// channels
	Emitter.Listen(Events.CHANNEL_EVENTS.CHANNEL_UPDATED, HandleChannelUpdated)
	Emitter.Listen(Events.CHANNEL_EVENTS.COMPULSORY_CHANNEL_CREATED, HandleCompulsoryChannelCreated)

	logger.Info("emitter listening to all events")
}

// users
func HandleUserCreated(data map[string]interface{}) {
	Emitter.Emit(Events.EMAIL_EVENTS.EMAIL_SENT, map[string]interface{}{
		"to":       data["email"].(string),
		"subject":  "Activate your Mize account",
		"template": "otp",
		"opts":     map[string]interface{}{"OTP": data["otp"], "HEADER": data["header"]},
	})
}
func HandleUserVerified(data map[string]string) {
	Emitter.Emit(Events.EMAIL_EVENTS.EMAIL_SENT, map[string]interface{}{
		"to":       data["email"],
		"subject":  "Welcome to Mize",
		"template": "welcome",
		"opts":     map[string]string{"FIRSTNAME": data["firstName"]},
	})
	eventsqueue.CreateAndEmitEvent(eventsqueue.CREATE_ACS_USER, map[string]interface{}{
		"id": data["id"],
	})
}
func HandleResendOtp(data map[string]interface{}) {
	Emitter.Emit(Events.EMAIL_EVENTS.EMAIL_SENT, map[string]interface{}{
		"to":       data["email"].(string),
		"subject":  "OTP request",
		"template": "otp",
		"opts":     map[string]interface{}{"OTP": data["otp"], "HEADER": data["header"]},
	})
}

// messages
func HandleMessageDeleted() {}

// sms
func HandleSMSSent(data map[string]interface{}) {
	eventsqueue.CreateAndEmitEvent(eventsqueue.SMS_REQUEST, map[string]interface{}{
		"to":      data["to"],
		"message": data["message"],
	})
}

// emails
func HandleEmailSent(data map[string]interface{}) {
	eventsqueue.CreateAndEmitEvent(eventsqueue.EMAIL_REQUEST, map[string]interface{}{
		"to":       data["to"],
		"subject":  data["subject"],
		"template": data["template"],
		"opts":     data["opts"],
	})
}

func HandleChannelUpdated(data map[string]interface{}) {
	eventsqueue.CreateAndEmitEvent(eventsqueue.CHANNEL_UPDATED, map[string]interface{}{
		"id":   data["id"],
		"data": data["data"],
	})
}

func HandleNotifyTaggedUsers(data map[string]interface{}) {
	eventsqueue.CreateAndEmitEvent(eventsqueue.NOTIFY_TAGGED, map[string]interface{}{
		"channel": data["channel"],
		"by":      data["by"],
		"msg":     data["msg"],
	})
}

func HandleCompulsoryChannelCreated(data map[string]interface{}) {
	eventsqueue.CreateAndEmitEvent(eventsqueue.EventTopic(eventsqueue.ADD_MEMBERS_TO_COMPULSORY_CHANNEL), map[string]interface{}{
		"channelId":   data["channelId"],
		"workspaceId": data["workspaceId"],
	})
}

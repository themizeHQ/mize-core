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
	Emitter.Listen(Events.MESSAGES_EVENTS.MESSAGE_SENT, HandleMessageSent)
	Emitter.Listen(Events.MESSAGES_EVENTS.MESSAGE_DELETED, HandleMessageDeleted)

	// sms
	Emitter.Listen(Events.SMS_EVENTS.SMS_SENT, HandleSMSSent)

	// emails
	Emitter.Listen(Events.EMAIL_EVENTS.EMAIL_SENT, HandleEmailSent)

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
func HandleMessageSent(data interface{}) {
	// realtime.CentrifugoController.Publish(fmt.Sprintf("%s-chat", data.(map[string]interface{})["to"]), data)
}
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

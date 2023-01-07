package emitter

import (
	"mize.app/emails"
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

	logger.Info("emitter listening to all events")
}

// users
func HandleUserCreated(data map[string]interface{}) {
	emails.SendEmail(data["email"].(string), "Activate your Mize account", "otp", map[string]interface{}{"OTP": data["otp"], "HEADER": data["header"]})
}
func HandleUserVerified(data map[string]string) {
	emails.SendEmail(data["email"], "Welcome to Mize", "welcome", map[string]string{"FIRSTNAME": data["firstName"]})
}
func HandleResendOtp(data map[string]interface{}) {
	emails.SendEmail(data["email"].(string), "OTP sent", "otp", map[string]interface{}{"OTP": data["otp"], "HEADER": data["header"]})
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

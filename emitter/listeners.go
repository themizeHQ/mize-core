package emitter

import (
	"fmt"

	"mize.app/emails"
	"mize.app/realtime"
)

func EmitterListener() {
	// auth
	Emitter.Listen(Events.AUTH_EVENTS.USER_CREATED, HandleUserCreated)
	Emitter.Listen(Events.AUTH_EVENTS.USER_VERIFIED, HandleUserVerified)
	Emitter.Listen(Events.AUTH_EVENTS.RESEND_OTP, HandleResendOtp)

	// messages
	Emitter.Listen(Events.MESSAGES_EVENTS.MESSAGE_SENT, HandleMessageSent)
	Emitter.Listen(Events.MESSAGES_EVENTS.MESSAGE_DELETED, HandleMessageDeleted)
}

// users
func HandleUserCreated(data map[string]string) {
	emails.SendEmail(data["email"], "Activate your Mize account", "otp", map[string]string{"OTP": data["otp"]})
}
func HandleUserVerified(data map[string]string) {
	emails.SendEmail(data["email"], "Welcome to Mize", "welcome", map[string]string{})
}
func HandleResendOtp(data map[string]interface{}) {
	emails.SendEmail(data["email"].(string), "Activate your Mize account", "otp", map[string]interface{}{"OTP": data["otp"]})
}

// messages
func HandleMessageSent(data interface{}) {
	realtime.CentrifugoController.Publish(fmt.Sprintf("%s-chat", data.(map[string]interface{})["to"]), data)
}
func HandleMessageDeleted() {}

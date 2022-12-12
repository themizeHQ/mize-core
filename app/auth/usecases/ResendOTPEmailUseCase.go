package usecases

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"mize.app/authentication"
	"mize.app/emitter"
	"mize.app/server_response"
)

func ResendOTPEmailUseCase(ctx *gin.Context, email string) error {
	otp, err := authentication.GenerateOTP(6)
	if err != nil {
		server_response.Response(ctx, http.StatusInternalServerError, "Something went wrong while generating otp", false, nil)
		return err
	}
	authentication.SaveOTP(ctx, email, otp, 5*time.Minute)
	emitter.Emitter.Emit(emitter.Events.AUTH_EVENTS.RESEND_OTP, map[string]interface{}{
		"email":  email,
		"otp":    strings.Split(otp, ""),
		"header": "OTP requested",
	})
	return nil
}

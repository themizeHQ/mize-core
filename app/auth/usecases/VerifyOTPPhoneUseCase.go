package usecases

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"mize.app/app/auth/types"
	userUseCases "mize.app/app/user/usecases/user"
	"mize.app/app_errors"
	"mize.app/authentication"
)

func VerifyOTPPhoneUsecase(ctx *gin.Context, payload types.VerifyPhoneData) error {
	valid, err := authentication.VerifyOTP(ctx, fmt.Sprintf("%s-otp", payload.Phone), payload.Otp)
	if err != nil {
		return err
	}
	if !valid {
		err = errors.New("wrong otp provided")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return err
	}
	success := userUseCases.UpdatePhoneUseCase(ctx, payload.Phone)
	if !success {
		err = errors.New("could not update phone number")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return err
	}
	return nil
}

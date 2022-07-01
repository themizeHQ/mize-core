package user

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	userModel "mize.app/app/user/models"
	useCases "mize.app/app/user/usecases/user"
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/emails"
	"mize.app/server_response"
)

func CacheUser(ctx *gin.Context) {
	var payload userModel.User
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	payload.RunHooks()
	err := useCases.CacheUserUseCase(ctx, &payload)
	if err != nil {
		return
	}
	otp, err := authentication.GenerateOTP(6)
	if err != nil {
		server_response.Response(ctx, http.StatusInternalServerError, "Something went wrong while creating user", false, nil)
		return
	}
	authentication.SaveOTP(ctx, payload.Email, otp, 5*time.Minute)
	emails.SendEmail(payload.Email, "Activate your Mize account", "otp", map[string]string{"OTP": string(otp)})
	server_response.Response(ctx, http.StatusCreated, "User created successfuly", true, nil)
}

func VerifyUser(ctx *gin.Context) {
	type VerifyData struct {
		Otp   string
		Email string
	}
	var payload VerifyData
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	valid, err := authentication.VerifyOTP(ctx, fmt.Sprintf("%s-otp", payload.Email), payload.Otp)
	if err != nil {
		return
	}
	if !valid {
		server_response.Response(ctx, http.StatusUnauthorized, "Wrong otp provided", false, nil)
		return
	}
	result, err := useCases.CreateUserUseCase(ctx, payload.Email)
	if err != nil {
		return
	}
	authentication.GenerateAccessToken(ctx, *result, payload.Email)
	authentication.GenerateRefreshToken(ctx, *result, payload.Email)
	emails.SendEmail(payload.Email, "Welcome to Mize", "welcome", map[string]string{})
	server_response.Response(ctx, http.StatusCreated, "Account verified", true, result)
}

func LoginUser(ctx *gin.Context) {
	type LoginDetails struct {
		Password string
		Account  string
	}
	var payload LoginDetails
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json"), StatusCode: http.StatusBadRequest})
		return
	}
	profile := useCases.LoginUserUseCase(ctx, payload.Account, payload.Password)
	if profile == nil {
		return
	}
	authentication.GenerateAccessToken(ctx, profile.Id.Hex(), profile.Email)
	authentication.GenerateRefreshToken(ctx, profile.Id.Hex(), profile.Email)
	server_response.Response(ctx, http.StatusCreated, "Login Successful", true, profile)
}

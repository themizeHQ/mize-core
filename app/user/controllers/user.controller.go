package user

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	userModel "mize.app/app/user/models"
	useCases "mize.app/app/user/usecases/user"
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/server_response"
)

func CacheUser(ctx *gin.Context) {
	var payload userModel.User
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, err)
		return
	}
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, err)
		return
	}
	response, err := useCases.CacheUserUseCase(ctx, &payload)
	if err != nil {
		return
	}
	otp, err := authentication.GenerateOTP(6)
	if err != nil {
		server_response.Response(ctx, http.StatusInternalServerError, "Something went wrong while creating user", false, nil)
		return
	}
	authentication.SaveOTP(ctx, payload.Email, otp, 5*time.Minute)
	server_response.Response(ctx, http.StatusCreated, "User created successfuly", true, response)
}

func VerifyUser(ctx *gin.Context) {
	type VerifyData struct {
		Otp   string
		Email string
	}
	var payload VerifyData
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, err)
		return
	}
	valid, err := authentication.VerifyOTP(ctx, fmt.Sprintf("%s-otp", payload.Email), payload.Otp)
	if err != nil {
		return
	}
	if !valid {
		ctx.Abort()
		server_response.Response(ctx, http.StatusUnauthorized, "Wrong otp provided", false, nil)
		return
	}
	result, err := useCases.CreateUserUseCase(ctx, payload.Email)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "Account verified", true, result)
}

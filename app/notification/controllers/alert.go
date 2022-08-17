package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	alertUseCases "mize.app/app/notification/usecases/alert"
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/server_response"
)

func SendAlert(ctx *gin.Context) {
	if ctx.GetString("Role") != string(authentication.ADMIN) {
		server_response.Response(ctx, http.StatusUnauthorized, "only admins can send alerts", false, nil)
		return
	}
	var payload alertUseCases.UserPayload
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the valid json"), StatusCode: http.StatusBadRequest})
		return
	}
	success := alertUseCases.CreateNewAlertUseCase(ctx, payload)
	if !success {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "alert sent", true, nil)
}

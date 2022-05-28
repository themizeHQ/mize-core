package application

import (
	"github.com/gin-gonic/gin"
	"net/http"

	application "mize.app/app/application/models"
	useCases "mize.app/app/application/usecases"
	"mize.app/app_errors"
	"mize.app/emails"
	"mize.app/server_response"
)

func CreateApplication(ctx *gin.Context) {
	var payload application.Application
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return
	}
	payload.Approved = false
	payload.Active = false
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return
	}
	result, err := useCases.CreateAppUseCase(ctx, payload)
	if err != nil {
		return
	}
	emails.SendEmail(payload.Email, "Your application has been created!", "app_created", map[string]string{"APP_NAME": payload.Name})
	server_response.Response(ctx, http.StatusCreated, "Application successfully created.", true, result)
}

package application

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/application/models"
	"mize.app/app/application/usecases"
	"mize.app/app_errors"
	"mize.app/emails"
	"mize.app/server_response"
)

func CreateApplication(ctx *gin.Context) {
	var payload models.Application
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	err := usecases.CreateAppUseCase(ctx, payload)
	if err != nil {
		return
	}
	emails.SendEmail(payload.Email, "Your application has been created!", "app_created", map[string]string{"APP_NAME": payload.Name})
	server_response.Response(ctx, http.StatusCreated, "Application successfully created.", true, nil)
}

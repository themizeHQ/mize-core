package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"mize.app/app/conversation/models"
	"mize.app/app/conversation/usecases/message"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func SendMessage(ctx *gin.Context) {
	var payload models.Message
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the required info"), StatusCode: http.StatusBadRequest})
		return
	}
	err := messages.SendMessageUseCase(ctx, payload)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "sent", true, nil)
}

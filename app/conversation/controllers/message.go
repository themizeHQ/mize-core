package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"mize.app/app/conversation/models"
	"mize.app/app/conversation/repository"
	"mize.app/app/conversation/usecases/message"
	"mize.app/app_errors"
	"mize.app/server_response"
	"mize.app/utils"
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

func FetchMessages(ctx *gin.Context) {
	to := ctx.Query("to")
	if to == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the conversation id"), StatusCode: http.StatusBadRequest})
		return
	}
	messageRepository := repository.GetMessageRepo()
	messages, err := messageRepository.FindMany(map[string]interface{}{
		"to":          utils.HexToMongoId(ctx, to),
		"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not fetch messages"), StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "messages fetched", true, messages)
}

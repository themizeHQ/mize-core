package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/workspace/models"
	channelUseCases "mize.app/app/workspace/usecases/channel"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func CreateChannel(ctx *gin.Context) {
	var payload []models.Channel
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json array value"), StatusCode: http.StatusBadGateway})
		return
	}
	id, err := channelUseCases.CreateChannelUseCase(ctx, payload)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "channel successfully created.", true, id)
}

func DeleteChannel(ctx *gin.Context) {
	channel_id := ctx.Params.ByName("id")
	success, err := channelUseCases.DeleteChannelUseCase(ctx, channel_id)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "channel deleted successfully", success, nil)
}

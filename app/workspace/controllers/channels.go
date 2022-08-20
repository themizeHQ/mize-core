package controllers

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"mize.app/app/workspace/models"
	channelUseCases "mize.app/app/workspace/usecases/channel"
	channelmembersUseCases "mize.app/app/workspace/usecases/channel_member"
	"mize.app/app_errors"
	"mize.app/server_response"
)

var waitGroup sync.WaitGroup

func CreateChannel(ctx *gin.Context) {
	var payload []models.Channel
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json array value"), StatusCode: http.StatusBadGateway})
		return
	}
	ids, err := channelUseCases.CreateChannelUseCase(ctx, payload)
	if err != nil {
		return
	}
	for _, id := range *ids {
		waitGroup.Add(1)
		go func(id string, w *sync.WaitGroup) {
			w.Done()
			channelmembersUseCases.CreateChannelMemberUseCase(ctx, id, true)
		}(id, &waitGroup)
	}
	server_response.Response(ctx, http.StatusCreated, "channel successfully created.", true, ids)
}

func DeleteChannel(ctx *gin.Context) {
	channel_id := ctx.Params.ByName("id")
	success, err := channelUseCases.DeleteChannelUseCase(ctx, channel_id)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "channel deleted successfully", success, nil)
}

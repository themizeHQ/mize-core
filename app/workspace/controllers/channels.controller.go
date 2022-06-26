package workspace

import (
	"net/http"

	"github.com/gin-gonic/gin"

	workspaceModel "mize.app/app/workspace/models"
	channelUseCases "mize.app/app/workspace/usecases/channel"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func CreateChannel(ctx *gin.Context) {
	var payload workspaceModel.Channel
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return
	}
	result, err := channelUseCases.CreateChannelUseCase(ctx, payload)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "Channel successfully created.", true, result)
}

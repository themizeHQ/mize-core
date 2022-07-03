package workspace

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	workspaceModel "mize.app/app/workspace/models"
	channelUseCases "mize.app/app/workspace/usecases/channel"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func CreateChannel(ctx *gin.Context) {
	var payload []workspaceModel.Channel
	workspace_id := ctx.Query("workspace_id")
	if workspace_id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a valid workspace id"), StatusCode: http.StatusBadGateway})
		return
	}
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json array value"), StatusCode: http.StatusBadGateway})
		return
	}
	id, err := channelUseCases.CreateChannelUseCase(ctx, workspace_id, payload)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "channel successfully created.", true, id)
}

package workspace

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	channelMemberUseCases "mize.app/app/workspace/usecases/channel_member"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func CreateChannelMember(ctx *gin.Context) {
	channel_id := ctx.Params.ByName("id")
	if channel_id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the valid json"), StatusCode: http.StatusBadGateway})
		return
	}
	id, err := channelMemberUseCases.CreateChannelMemberUseCase(ctx, channel_id)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "joined channel successfully", true, id)
}

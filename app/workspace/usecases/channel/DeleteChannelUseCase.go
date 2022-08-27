package channel

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/constants/channel"
	"mize.app/utils"
)

func DeleteChannelUseCase(ctx *gin.Context, channel_id string) (bool, error) {
	channelMemberRepo := repository.GetChannelMemberRepo()
	channelMembership, err := channelMemberRepo.FindOneByFilter(map[string]interface{}{
		"userId":      utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	})
	if err != nil {
		err = errors.New("workspace does not exist")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
		return false, err
	}
	if !channelMembership.Admin {
		err = errors.New("only admins can delete channels")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return false, err
	}
	has_access := channelMembership.HasAccess(channelMembership.AdminAccess,
		[]channel.ChannelAdminAccess{channel.CHANNEL_FULL_ACCESS, channel.CHANNEL_DELETE_ACCESS})
	if !has_access {
		err = errors.New("you do not have delete access to this channel")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return false, err
	}
	return channelMemberRepo.DeleteById(channel_id)
}

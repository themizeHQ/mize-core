package channel_member

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func LeaveChannelUseCase(ctx *gin.Context, id *string) bool {
	channelMemberRepo := repository.GetChannelMemberRepo()
	exists, err := channelMemberRepo.CountDocs(map[string]interface{}{
		"_id":         *id,
		"userId":      *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong"), StatusCode: http.StatusInternalServerError})
		return false
	}
	if exists == 0 {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you are not a member of this channel"), StatusCode: http.StatusUnauthorized})
		return false
	}
	deleted, err := channelMemberRepo.DeleteById(*id)
	if err != nil || !deleted {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not leave channel"), StatusCode: http.StatusInternalServerError})
		return false
	}
	return deleted
}

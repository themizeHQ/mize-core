package channel_member

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func PinChannelMember(ctx *gin.Context, id *string) bool {
	channelMemberRepo := repository.GetChannelMemberRepo()
	exists, err := channelMemberRepo.CountDocs(map[string]interface{}{
		"userId": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"_id":    *id,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong"), StatusCode: http.StatusInternalServerError})
		return false
	}
	if exists == 0 {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you are not a member of this channel"), StatusCode: http.StatusUnauthorized})
		return false
	}
	pinnedCount, err := channelMemberRepo.CountDocs(map[string]interface{}{
		"userId":      utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		"pinned":      true,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong"), StatusCode: http.StatusInternalServerError})
		return false
	}
	if pinnedCount >= 5 {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you can pin only 5 channels"), StatusCode: http.StatusUnauthorized})
		return false
	}
	success, err := channelMemberRepo.UpdatePartialById(ctx, *id, map[string]interface{}{
		"pinned": true,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong"), StatusCode: http.StatusInternalServerError})
		return false
	}
	return success
}

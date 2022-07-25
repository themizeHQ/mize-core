package workspace

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	workspace "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func CreateChannelMemberUseCase(ctx *gin.Context, channel_id string) (*string, error) {
	workspaceMemberRepo := workspaceRepo.GetChannelMemberRepo()
	member := workspace.ChannelMember{
		ChannelId:   *utils.HexToMongoId(ctx, channel_id),
		WorkspaceId: *utils.HexToMongoId(ctx, ctx.GetString("Workspace")), Username: ctx.GetString("Username"),
		UserId:   *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		JoinDate: primitive.NewDateTimeFromTime(time.Now()),
	}
	if err := member.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	id, err := workspaceMemberRepo.CreateOne(member)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	return id, nil
}

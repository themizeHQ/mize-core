package channel_member

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"mize.app/app/workspace/models"
	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func CreateChannelMemberUseCase(ctx *gin.Context, channel_id string) (*string, error) {
	channelMemberRepo := repository.GetChannelMemberRepo()
	channelRepo := repository.GetChannelRepo()
	channel, err := channelRepo.FindById(channel_id)
	if channel == nil {
		err = errors.New("channel does not exist")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	if channel.WorkspaceId != *utils.HexToMongoId(ctx, ctx.GetString("Workspace")) {
		err = errors.New("you do not have access to this channel")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	exists, err := channelMemberRepo.CountDocs(map[string]interface{}{
		"userId":    *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"channelId": *utils.HexToMongoId(ctx, channel_id),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	if exists > 0 {
		err = errors.New("you are already a member of this channel")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	member := models.ChannelMember{
		ChannelId:   *utils.HexToMongoId(ctx, channel_id),
		WorkspaceId: *utils.HexToMongoId(ctx, ctx.GetString("Workspace")), Username: ctx.GetString("Username"),
		UserId:   *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		JoinDate: primitive.NewDateTimeFromTime(time.Now()),
	}
	if err := member.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	id, err := channelMemberRepo.CreateOne(member)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	return id, nil
}

package channel_member

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/workspace/models"
	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	channel_constants "mize.app/constants/channel"
	"mize.app/utils"
)

func CreateChannelMemberUseCase(ctx *gin.Context, channel_id string, name *string, admin bool) (*string, error) {
	channelMemberRepo := repository.GetChannelMemberRepo()
	channelRepo := repository.GetChannelRepo()
	channel, err := channelRepo.FindById(channel_id, options.FindOne().SetProjection(map[string]int{
		"workspaceId": 1,
		"name":        1,
	}))
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
	if name == nil {
		name = &channel.Name
	}
	member := models.ChannelMember{
		ChannelId:   *utils.HexToMongoId(ctx, channel_id),
		WorkspaceId: *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		Username:    ctx.GetString("Username"),
		UserId:      *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		LastSent:    primitive.NewDateTimeFromTime(time.Now()),
		Admin:       admin,
		ChannelName: *name,
		AdminAccess: func() []channel_constants.ChannelAdminAccess {
			if admin {
				return []channel_constants.ChannelAdminAccess{channel_constants.CHANNEL_FULL_ACCESS}
			}
			return []channel_constants.ChannelAdminAccess{}
		}(),
		JoinDate: primitive.NewDateTimeFromTime(time.Now()),
	}
	if err := member.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	m, err := channelMemberRepo.CreateOne(member)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	id := m.Id.Hex()
	return &id, nil
}

package channel_member

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"mize.app/app/workspace/models"
	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	channel_constants "mize.app/constants/channel"
	"mize.app/logger"
	"mize.app/utils"
)

func CreateChannelMemberUseCase(ctx *gin.Context, channel_id string, name *string, created bool) (*string, error) {
	channelMemberRepo := repository.GetChannelMemberRepo()
	channelRepo := repository.GetChannelRepo()
	channel, err := channelRepo.FindOneByFilter(map[string]interface{}{
		"_id":         channel_id,
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	}, options.FindOne().SetProjection(map[string]int{
		"workspaceId": 1,
		"name":        1,
		"private":     1,
	}))
	if channel == nil {
		err = errors.New("channel does not exist")
		if !created {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		}
		return nil, err
	}
	if err != nil {
		logger.Error(errors.New("could not create channel member"), zap.Error(err))
		if !created {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		}
		return nil, err
	}
	if channel.Private && !created && !ctx.GetBool("Admin") {
		err = errors.New("you do not have access to this private channel")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return nil, err
	}
	exists, err := channelMemberRepo.CountDocs(map[string]interface{}{
		"userId":    *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"channelId": *utils.HexToMongoId(ctx, channel_id),
	})
	if err != nil {
		logger.Error(errors.New("could not create channel member"), zap.Error(err))
		if !created {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		}
		return nil, err
	}
	if exists != 0 {
		err = errors.New("you are already a member of this channel")
		if !created {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		}
		return nil, err
	}
	if name == nil {
		name = &channel.Name
	}
	workspaceMemberRepo := repository.GetWorkspaceMember()
	workspaceMember, err := workspaceMemberRepo.FindOneByFilter(map[string]interface{}{
		"userId":      *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"workspaceId": channel.WorkspaceId,
	}, options.FindOne().SetProjection(map[string]interface{}{
		"profileImageThumbnail": 1,
		"profileImage":          1,
	}))
	if err != nil {
		logger.Error(errors.New("could not create channel member"), zap.Error(err))
		if !created {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		}
		return nil, err
	}
	if workspaceMember == nil {
		logger.Error(errors.New("could not create channel member"), zap.Error(err))
		if !created {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		}
		return nil, err
	}
	member := models.ChannelMember{
		ChannelId:   *utils.HexToMongoId(ctx, channel_id),
		WorkspaceId: *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		Username:    ctx.GetString("Username"),
		UserId:      *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		LastSent:    primitive.NewDateTimeFromTime(time.Now()),
		Admin:       ctx.GetBool("Admin"),
		ChannelName: *name,
		AdminAccess: func() []channel_constants.ChannelAdminAccess {
			if !created {
				return []channel_constants.ChannelAdminAccess{channel_constants.CHANNEL_FULL_ACCESS}
			}
			return []channel_constants.ChannelAdminAccess{}
		}(),
		JoinDate:              primitive.NewDateTimeFromTime(time.Now()),
		ProfileImage:          workspaceMember.ProfileImage,
		ProfileImageThumbNail: *workspaceMember.ProfileImageThumbNail,
	}
	if err := member.Validate(); err != nil {
		logger.Error(errors.New("could not create channel member"), zap.Error(err))
		if !created {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		}
		return nil, err
	}
	m, err := channelMemberRepo.CreateOne(member)
	if err != nil {
		logger.Error(errors.New("could not create channel member"), zap.Error(err))
		if !created {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		}
		return nil, err
	}
	id := m.Id.Hex()
	return &id, nil
}

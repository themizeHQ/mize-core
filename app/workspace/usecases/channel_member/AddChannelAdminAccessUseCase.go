package channel_member

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	channel_member_constants "mize.app/constants/channel"
	"mize.app/logger"
)

func AddChannelAdminAccessUseCase(ctx *gin.Context, id string, channel_id string, permissions []channel_member_constants.ChannelAdminAccess) error {
	channelMemberRepo := repository.GetChannelMemberRepo()
	channelRepo := repository.GetChannelRepo()
	workspaceRepo := repository.GetWorkspaceRepo()
	var wg sync.WaitGroup

	member, err := channelMemberRepo.FindOneByFilter(map[string]interface{}{
		"_id": id,
	}, options.FindOne().SetProjection(map[string]interface{}{
		"userId": 1,
	}))
	if err != nil {
		e := errors.New("error checking if a user exists")
		logger.Error(e, zap.Error(err))
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: e, StatusCode: http.StatusInternalServerError})
		return e
	}
	if member == nil {
		e := errors.New("user is not a member of this channel")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: e, StatusCode: http.StatusInternalServerError})
		return e
	}

	var chanErr1 = make(chan error)
	wg.Add(1)
	go func(e chan error) {
		defer func() {
			wg.Done()
		}()
		channel, err := channelRepo.FindOneByFilter(map[string]interface{}{
			"_id": channel_id,
		}, options.FindOne().SetProjection(map[string]interface{}{
			"createdBy": 1,
		}))
		if err != nil {
			e2 := errors.New("error fetching workspace")
			logger.Error(e2, zap.Error(err))
			e <- err
			return
		}
		if channel == nil {
			e2 := errors.New("channel does not exist")
			e <- e2
			return
		}
		if channel.CreatedBy.Hex() == member.UserId.Hex() {
			e2 := errors.New("you cannot modify permissions for this user")
			e <- e2
			return
		}
		e <- nil
	}(chanErr1)

	var chanErr2 = make(chan error)
	wg.Add(1)
	go func(e chan error) {
		defer func() {
			wg.Done()
		}()
		workspace, err := workspaceRepo.FindOneByFilter(map[string]interface{}{
			"_id": ctx.GetString("Workspace"),
		}, options.FindOne().SetProjection(map[string]interface{}{
			"createdBy": 1,
		}))
		if err != nil {
			e2 := errors.New("error fetching workspace")
			logger.Error(e2, zap.Error(err))
			e <- err
			return
		}
		if workspace == nil {
			e2 := errors.New("workspace does not exist")
			e <- e2
			return
		}
		if workspace.CreatedBy.Hex() == member.UserId.Hex() {
			e2 := errors.New("you cannot modify permissions for this user")
			e <- e2
			return
		}
		e <- nil
	}(chanErr2)

	err1 := <-chanErr1
	err2 := <-chanErr2

	if err1 != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err1, StatusCode: http.StatusUnauthorized})
		return err1
	}

	if err2 != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err2, StatusCode: http.StatusUnauthorized})
		return err2
	}
	wg.Wait()
	success, err := channelMemberRepo.UpdatePartialById(ctx, id, map[string]interface{}{
		"admin":       len(permissions) != 0,
		"adminAccess": permissions,
	})
	if err != nil || !success {
		e := errors.New("could not update channel admin priviledges")
		logger.Error(e, zap.Error(err))
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: e, StatusCode: http.StatusNotFound})
		return e
	}
	return nil
}

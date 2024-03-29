package workspace_invite

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	notificationUseCases "mize.app/app/notification/usecases"
	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func RejectWorkspaceInviteUseCase(ctx *gin.Context, workspace_invite_id string) (bool, error) {
	var workspaceInviteRepoInstance = repository.GetWorkspaceInviteRepo()
	invite, err := workspaceInviteRepoInstance.FindOneByFilter(map[string]interface{}{
		"_id":   *utils.HexToMongoId(ctx, workspace_invite_id),
		"email": ctx.GetString("Email"),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return false, err
	}
	if invite == nil {
		err = errors.New("invite does not exist")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return false, err
	}
	if !invite.Success {
		err = errors.New("invite was not successfully sent")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return false, err
	}
	if invite.Accepted != nil {
		if *invite.Accepted {
			err = errors.New("invite has already been accepted")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
			return false, err
		}
		if !*invite.Accepted {
			err = errors.New("invite has already been rejeted")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
			return false, err
		}
	}
	state := false
	invite.Accepted = &state
	success, err := workspaceInviteRepoInstance.UpdateById(ctx, invite.Id.Hex(), invite)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return false, err
	}
	if !success {
		err = errors.New("something went wrong while rejecting your invite")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return false, err
	}
	notificationUseCases.UpdateNotificationUseCase(ctx, map[string]interface{}{
		"reacted": true,
	})
	return true, nil
}

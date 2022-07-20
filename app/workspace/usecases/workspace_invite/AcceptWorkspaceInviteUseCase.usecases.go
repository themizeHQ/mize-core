package workspace

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
)

func AcceptWorkspaceInviteUseCase(ctx *gin.Context, workspace_invite_id string) (*string, error) {
	var workspaceInviteRepoInstance = workspaceRepo.GetWorkspaceInviteRepo()
	invite, err := workspaceInviteRepoInstance.FindById(workspace_invite_id)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	if invite == nil {
		err = errors.New("invite does not exist")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	if !invite.Success {
		err = errors.New("invite was not successfully sent")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	if invite.Accepted != nil {
		if *invite.Accepted {
			err = errors.New("invite has already been accepted")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
			return nil, err
		}
		if !*invite.Accepted {
			err = errors.New("invite has already been rejeted")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
			return nil, err
		}
	}
	state := true
	invite.Accepted = &state
	success, err := workspaceInviteRepoInstance.UpdateById(ctx, invite.Id.Hex(), invite)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	if !success {
		err = errors.New("something went wrong while accepting your invite")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	workspaceId := invite.WorkspaceId.Hex()
	return &workspaceId, nil
}

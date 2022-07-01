package workspace

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
)

func CreateChannelUseCase(ctx *gin.Context, payload workspaceModel.Channel) (*string, error) {
	var channelRepoInstance = workspaceRepo.GetChannelRepo()
	var workspaceRepoInstance = workspaceRepo.GetWorkspaceRepo()
	workspace, err := workspaceRepoInstance.FindById(payload.WorkspaceId)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("workspace does not exist"), StatusCode: http.StatusNotFound})
		return nil, err
	}
	has_access := false
	if workspace.CreatedBy == ctx.GetString("UserId") {
		has_access = true
	} else {
		for _, admin := range workspace.Admins {
			if admin == ctx.GetString("UserId") {
				has_access = true
				break
			}
		}
	}
	if !has_access {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("only admins can create channels"), StatusCode: http.StatusUnauthorized})
		return nil, err
	}
	payload.CreatedBy = ctx.GetString("UserId")
	id, err := channelRepoInstance.CreateOne(payload)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("channel creation failed"), StatusCode: http.StatusInternalServerError})
		return nil, err
	}
	return id, nil
}

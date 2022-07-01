package workspace

import (
	"net/http"

	"github.com/gin-gonic/gin"

	workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
)

func CreateWorkspaceInviteUseCase(ctx *gin.Context, payload workspaceModel.WorkspaceInvite) error {
	var workspaceInviteRepoInstance = workspaceRepo.GetWorkspaceInviteRepo()
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	workspaceInviteRepoInstance.CreateOne(payload)
	return nil
}

package workspace

import (
	"net/http"

	"github.com/gin-gonic/gin"

	workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
)

func CreateWorkspaceInviteUseCase(ctx *gin.Context, payload workspaceModel.WorkspaceInvite) bool {
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return false
	}
	workspaceRepo.WorkspaceInviteRepository.CreateOne(ctx, &payload)
	return true
}

package workspace

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
)

func CreateWorkspaceUseCase(ctx *gin.Context, payload workspaceModel.Workspace) (*string, error) {
	var workspaceRepoInstance = workspaceRepo.GetWorkspaceRepo()
	payload.CreatedBy = ctx.GetString("UserId")
	payload.Email = ctx.GetString("Email")
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, err
	}
	id, err := workspaceRepoInstance.CreateOne(payload)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("workspace creation failed"), StatusCode: http.StatusInternalServerError})
		return nil, err
	}
	return id, nil
}

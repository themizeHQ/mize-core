package workspace

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	userRepo "mize.app/app/user/repository"
	workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
)

func CreateWorkspaceUseCase(ctx *gin.Context, payload workspaceModel.Workspace) (*string, error) {
	user := userRepo.UserRepository.FindOneByFilter(ctx, map[string]interface{}{"_id": ctx.GetString("UserId")})
	if user == nil {
		err := errors.New("user does not exist")
		app_errors.ErrorHandler(ctx, err, http.StatusNotFound)
		return nil, err
	}
	payload.Email = user.Email
	payload.CreatedBy = user.Id
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return nil, err
	}
	response := workspaceRepo.WorkspaceRepository.CreateOne(ctx, &payload)
	if response == nil {
		err := errors.New("workspace creation failed")
		app_errors.ErrorHandler(ctx, err, http.StatusNotFound)
		return nil, err
	}
	return response, nil
}

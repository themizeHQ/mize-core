package workspace

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/workspace/models"
	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func CreateWorkspaceUseCase(ctx *gin.Context, payload models.Workspace) (name *string, id *string, err error) {
	var workspaceRepoInstance = repository.GetWorkspaceRepo()
	payload.CreatedBy = *utils.HexToMongoId(ctx, ctx.GetString("UserId"))
	payload.Email = ctx.GetString("Email")
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return nil, nil, err
	}
	payload.MemberCount = 1
	workspace, err := workspaceRepoInstance.CreateOne(payload)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("workspace creation failed"), StatusCode: http.StatusInternalServerError})
		return nil, nil, err
	}
	wid := workspace.Id.Hex()
	return &workspace.Name, &wid, nil
}

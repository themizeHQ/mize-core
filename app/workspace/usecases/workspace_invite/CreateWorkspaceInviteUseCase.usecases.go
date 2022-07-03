package workspace

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
)

func CreateWorkspaceInviteUseCase(ctx *gin.Context, filter map[string]interface{}, payload map[string]interface{}) error {
	var workspaceInviteRepoInstance = workspaceRepo.GetWorkspaceInviteRepo()
	if err := interface{}(payload["$set"]).(*workspaceModel.WorkspaceInvite).Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	workspaceInviteRepoInstance.UpdateByField(filter, payload, options.Update().SetUpsert(true))
	return nil
}

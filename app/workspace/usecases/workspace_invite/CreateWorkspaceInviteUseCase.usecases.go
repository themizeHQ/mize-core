package workspace

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
)

func CreateWorkspaceInviteUseCase(ctx *gin.Context, filter map[string]interface{}, payload *workspaceModel.WorkspaceInvite) error {
	var workspaceInviteRepoInstance = workspaceRepo.GetWorkspaceInviteRepo()
	if err := payload.Validate(); err != nil {
		return err
	}
	workspaceInviteRepoInstance.UpdateOrCreateByField(filter, payload, options.Update().SetUpsert(true))
	return nil
}

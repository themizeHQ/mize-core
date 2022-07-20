package workspace

import (
	"time"

	"github.com/gin-gonic/gin"
	workspace "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/utils"
)

func CreateWorkspaceMemberUseCase(ctx *gin.Context, workspace_id string, admin bool) (*string, error) {
	workspaceMemberRepo := workspaceRepo.GetWorkspaceMember()
	workspace_id_hex := *utils.HexToMongoId(ctx, workspace_id)
	payload := workspace.WorkspaceMember{WorkspaceId: workspace_id_hex,
		Username: ctx.GetString("Username"), UserId: *utils.HexToMongoId(ctx, ctx.GetString("UserId")), Admin: admin, JoinDate: time.Now().Unix()}
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	return workspaceMemberRepo.CreateOne(payload)
}

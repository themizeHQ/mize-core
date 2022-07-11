package workspace

import (
	"time"

	"github.com/gin-gonic/gin"
	workspace "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
)

func CreateWorkspaceMemberUseCase(ctx *gin.Context, workspace_id string, admin bool) (*string, error) {
	workspaceMemberRepo := workspaceRepo.GetWorkspaceMember()
	payload := workspace.WorkspaceMember{WorkspaceId: workspace_id,
		Username: ctx.GetString("Username"), UserId: ctx.GetString("UserId"), Admin: admin, JoinDate: time.Now().Unix()}
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	return workspaceMemberRepo.CreateOne(payload)
}

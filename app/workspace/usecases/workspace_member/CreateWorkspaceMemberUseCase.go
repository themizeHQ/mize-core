package workspace_member

import (
	"time"

	"github.com/gin-gonic/gin"
	"mize.app/app/workspace/models"
	"mize.app/app/workspace/repository"
	workspace_constants "mize.app/constants/workspace"
	"mize.app/utils"
)

func CreateWorkspaceMemberUseCase(ctx *gin.Context, workspace_id *string, name *string, admin bool) (*string, error) {
	workspaceMemberRepo := repository.GetWorkspaceMember()
	workspace_id_hex := *utils.HexToMongoId(ctx, *workspace_id)
	payload := models.WorkspaceMember{
		WorkspaceId:   workspace_id_hex,
		Username:      ctx.GetString("Username"),
		WorkspaceName: *name,
		UserId:        *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		Admin:         admin,
		AdminAccess: func() []workspace_constants.AdminAccessType {
			if admin {
				return []workspace_constants.AdminAccessType{workspace_constants.AdminAccess.FULL_ACCESS}
			}
			return []workspace_constants.AdminAccessType{}
		}(),
		JoinDate: time.Now().Unix()}
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	member, err := workspaceMemberRepo.CreateOne(payload)
	id := member.Id.Hex()
	return &id, err
}

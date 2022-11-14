package teammembers

import (
	"github.com/gin-gonic/gin"

	teamsRepo "mize.app/app/teams/repository"
	types "mize.app/app/teams/types"
	"mize.app/utils"
)

func RemoveTeamMembers(ctx *gin.Context, ids types.IDArray) {
	teamMemberRepo := teamsRepo.GetTeamMemberRepo()
	teamMemberRepo.DeleteMany(ctx, map[string]interface{}{
		"_id": map[string]interface{}{
			"$in": ids,
		},
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	})
}

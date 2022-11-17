package teamactivity

import (
	"github.com/gin-gonic/gin"
	"mize.app/app/teams/models"
	"mize.app/app/teams/repository"
	"mize.app/utils"
)

func CreateTeamActivityUseCase(ctx *gin.Context, payload models.TeamActivity) {
	payload.WorkspaceId = *utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
	teamActivityRepository := repository.GetTeamActivityRepo()
	teamActivityRepository.CreateOne(payload)
}

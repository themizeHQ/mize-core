package channel

import (
	"github.com/gin-gonic/gin"
	"mize.app/app/auth"
	workspaceRepository "mize.app/app/workspace/repository"
)

func DeleteChannelUseCase(ctx *gin.Context, channel_id string) (bool, error) {
	success, err := auth.HasChannelAccess(ctx, channel_id)
	if !success || err != nil {
		return success, err
	}
	channelMemberRepo := workspaceRepository.GetChannelMemberRepo()
	return channelMemberRepo.DeleteById(channel_id)
}

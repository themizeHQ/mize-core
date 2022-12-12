package channel

import (
	"github.com/gin-gonic/gin"
	workspaceRepository "mize.app/app/workspace/repository"
	"mize.app/authentication"
	channelConstants "mize.app/constants/channel"
)

func DeleteChannelUseCase(ctx *gin.Context, channel_id string) (bool, error) {
	success, err := authentication.HasChannelAccess(ctx, channel_id, []channelConstants.ChannelAdminAccess{channelConstants.CHANNEL_DELETE_ACCESS})
	if !success || err != nil {
		return success, err
	}
	channelMemberRepo := workspaceRepository.GetChannelMemberRepo()
	return channelMemberRepo.DeleteById(channel_id)
}

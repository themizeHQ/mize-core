package channel

import (
	"github.com/gin-gonic/gin"
	"mize.app/app/workspace"
	"mize.app/authentication"
	channelConstants "mize.app/constants/channel"
	"mize.app/emitter"
)

func UpdateChannelInfoUseCase(ctx *gin.Context, id string, payload workspace.UpdateChannelType) error {
	success, err := authentication.HasChannelAccess(ctx, id, []channelConstants.ChannelAdminAccess{channelConstants.CHANNEL_INFO_EDIT_ACCESS, channelConstants.CHANNEL_FULL_ACCESS})
	if !success || err != nil {
		return err
	}
	emitter.Emitter.Emit(emitter.Events.CHANNEL_EVENTS.CHANNEL_UPDATED, map[string]interface{}{
		"id":   id,
		"data": payload,
	})
	return nil
}

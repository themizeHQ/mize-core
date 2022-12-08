package conversation

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"mize.app/app/conversation/models"
	"mize.app/app/conversation/repository"
	"mize.app/app/conversation/types"
	"mize.app/app_errors"
	"mize.app/realtime"
	"mize.app/utils"
)

func CreateConversationMemberUseCase(ctx *gin.Context, payload *types.CreateConv, convId *primitive.ObjectID, recipientName string, userId string, reciepientId primitive.ObjectID, profileImage *string) (*string, error) {
	member := models.ConversationMember{
		ConversationId: *convId,
		UserId:         *utils.HexToMongoId(ctx, userId),
		ReciepientName: recipientName,
		ReciepientId:   reciepientId,
		ProfileImage: func() string {
			if profileImage == nil {
				return ""
			}
			return *profileImage
		}(),
		WorkspaceId: func() *primitive.ObjectID {
			if ctx.GetString("Workspace") == "" {
				return nil
			}
			return utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
		}(),
		WorkspaceConv: func() bool {
			return ctx.GetString("Workspace") != ""
		}(),
	}
	convMemberRepo := repository.GetConversationMemberRepo()
	data, err := convMemberRepo.CreateOne(member)
	if err != nil {
		err = errors.New("could not create conversation")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return nil, err
	}
	id := data.Id.Hex()
	realtime.CentrifugoController.Publish(fmt.Sprintf("%s-new-conversation", payload.ReciepientId.Hex()), map[string]interface{}{
		"userName":       member.ReciepientId.Hex(),
		"conversationId": member.ConversationId.Hex(),
		"profileImage":   member.ProfileImage,
	})
	return &id, err
}

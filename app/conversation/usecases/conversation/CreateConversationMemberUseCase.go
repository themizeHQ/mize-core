package conversation

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"mize.app/app/conversation/models"
	"mize.app/app/conversation/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func CreateConversationMemberUseCase(ctx *gin.Context, convId *primitive.ObjectID, recipientName string, userId string, profileImage *string, profileImageThumbnail *string) (*string, error) {
	member := models.ConversationMember{
		ConversationId:        *convId,
		UserId:                *utils.HexToMongoId(ctx, userId),
		ReciepientName:        recipientName,
		ReciepientId:          *utils.HexToMongoId(ctx, userId),
		ProfileImage:          *profileImage,
		ProfileImageThumbNail: *profileImageThumbnail,
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
	return &id, err
}

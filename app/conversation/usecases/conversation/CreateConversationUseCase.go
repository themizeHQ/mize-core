package conversation

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/conversation/models"
	"mize.app/app/conversation/repository"
	"mize.app/app/conversation/types"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func CreateConversationUseCase(ctx *gin.Context, payload *types.CreateConv) (*primitive.ObjectID, *string, *string, error) {
	convRepo := repository.GetConversationRepo()
	workspaceConv := func() bool {
		return ctx.GetString("Workspace") != ""
	}()
	var reciepientName string
	var profileImage *string
	if workspaceConv {
		workspaceMemberRepo := workspaceRepo.GetWorkspaceMember()
		member, err := workspaceMemberRepo.FindById(payload.ReciepientId.Hex(), options.FindOne().SetProjection(map[string]int{
			"userName":     1,
			"profileImage": 1,
		}))
		if err != nil {
			err = errors.New("could not create conversation")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return nil, nil, nil, err
		}
		if member == nil {
			err = errors.New("user is not a member of this workspace")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
			return nil, nil, nil, err
		}
		reciepientName = member.Username
		profileImage = member.ProfileImage
	}
	exists, err := convRepo.CountDocs(map[string]interface{}{
		"participants": []primitive.ObjectID{payload.ReciepientId, *utils.HexToMongoId(ctx, ctx.GetString("UserId"))},
	})
	if err != nil {
		err = errors.New("could not create conversation")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return nil, nil, nil, err
	}
	if exists != 0 {
		err = errors.New("conversation already exists")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
		return nil, nil, nil, err
	}
	conversationData := models.Conversation{
		Participants:  []primitive.ObjectID{payload.ReciepientId, *utils.HexToMongoId(ctx, ctx.GetString("UserId"))},
		WorkspaceConv: workspaceConv,
		WorkspaceId: func() *primitive.ObjectID {
			if ctx.GetString("Workspace") == "" {
				return nil
			}
			return utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
		}(),
	}
	conv, err := convRepo.CreateOne(conversationData)
	if err != nil {
		err = errors.New("could not create conversation")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return nil, nil, nil, err
	}
	convId := conv.Id
	return &convId, &reciepientName, profileImage, err
}

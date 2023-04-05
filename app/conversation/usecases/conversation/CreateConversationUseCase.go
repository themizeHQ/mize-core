package conversation

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/conversation/models"
	"mize.app/app/conversation/repository"
	userRepo "mize.app/app/user/repository"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func CreateConversationUseCase(ctx *gin.Context, reciepientId string) (*primitive.ObjectID, *string, *string, *string, error) {
	convRepo := repository.GetConversationRepo()
	workspaceConv := ctx.GetString("Workspace") != ""
	var reciepientName string
	var profileImage *string
	var profileImageThumbnail *string
	if workspaceConv {
		workspaceMemberRepo := workspaceRepo.GetWorkspaceMember()
		member, err := workspaceMemberRepo.FindOneByFilter(map[string]interface{}{
			"userId":      *utils.HexToMongoId(ctx, reciepientId),
			"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		}, options.FindOne().SetProjection(map[string]int{
			"userName":              1,
			"profileImage":          1,
			"profileImageThumbnail": 1,
		}))
		if err != nil {
			err = errors.New("could not create conversation")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return nil, nil, nil, nil, err
		}
		if member == nil {
			err = errors.New("user is not a member of this workspace")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
			return nil, nil, nil, nil, err
		}
		reciepientName = member.Username
		profileImage = member.ProfileImage
		profileImageThumbnail = member.ProfileImageThumbNail
	} else {
		userRepo := userRepo.GetUserRepo()
		user, err := userRepo.FindById(reciepientId, options.FindOne().SetProjection(map[string]int{
			"userName":              1,
			"profileImage":          1,
			"profileImageThumbnail": 1,
		}))
		if err != nil {
			err = errors.New("could not create conversation")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return nil, nil, nil, nil, err
		}
		if user == nil {
			err = errors.New("user does not exist")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
			return nil, nil, nil, nil, err
		}
		reciepientName = user.UserName
		profileImage = user.ProfileImage
		profileImageThumbnail = user.ProfileImageThumbNail
	}
	exists, err := convRepo.CountDocs(map[string]interface{}{
		"workspaceConv": workspaceConv,
		"participants": []primitive.ObjectID{*utils.HexToMongoId(ctx, reciepientId), *utils.HexToMongoId(ctx, ctx.GetString("UserId"))},
	})
	if err != nil {
		err = errors.New("could not create conversation")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return nil, nil, nil, nil, err
	}
	if exists != 0 {
		err = errors.New("conversation already exists")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
		return nil, nil, nil, nil, err
	}
	conversationData := models.Conversation{
		Participants:  []primitive.ObjectID{*utils.HexToMongoId(ctx, reciepientId), *utils.HexToMongoId(ctx, ctx.GetString("UserId"))},
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
		return nil, nil, nil, nil, err
	}
	convId := conv.Id
	return &convId, &reciepientName, profileImage, profileImageThumbnail, err
}

package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/conversation/repository"
	convUseCases "mize.app/app/conversation/usecases/conversation"
	"mize.app/app_errors"
	"mize.app/server_response"
	"mize.app/utils"
)

func FetchConversation(ctx *gin.Context) {
	convRepo := repository.GetConversationMemberRepo()

	startID := ctx.Query("id")
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 15
	}
	conversations, err := convRepo.FindManyStripped(map[string]interface{}{
		"workspaceId": func() *primitive.ObjectID {
			if ctx.GetString("Workspace") == "" {
				return nil
			}
			return utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
		}(),
		"_id": func() map[string]interface{} {
			if startID == "" {
				return map[string]interface{}{"$gt": primitive.NilObjectID}
			}
			return map[string]interface{}{"$lt": *utils.HexToMongoId(ctx, startID)}
		}(),
		"userId": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
	}, options.Find().SetSort(map[string]interface{}{
		"lastMessageSent": -1,
	}), options.Find().SetSort(map[string]interface{}{
		"_id": -1,
	}),&options.FindOptions{
		Limit: &limit,
	},  options.Find().SetProjection(
		map[string]interface{}{
			"lastMessage":           1,
			"unreadMessages":        1,
			"reciepientId":          1,
			"reciepientName":        1,
			"lastMessageSent":       1,
			"profileImage":          1,
			"profileImageThumbnail": 1,
			"conversationId":        1,
		},
	))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in message data"), StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "conversations fetched", true, conversations)
}

func SearchWorkspaceDM(ctx *gin.Context) {
	term := ctx.Query("term")
	convRepo := repository.GetConversationMemberRepo()

	conversations, err := convRepo.FindManyStripped(map[string]interface{}{
		"workspaceId":    *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		"userId":         *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"reciepientName": map[string]interface{}{"$regex": term, "$options": "im"},
	}, options.Find().SetProjection(
		map[string]interface{}{
			"lastMessage":     1,
			"unreadMessages":  1,
			"reciepientId":    1,
			"reciepientName":  1,
			"lastMessageSent": 1,
			"profileImage":    1,
			"conversationId":  1,
		},
	))

	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in message data"), StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "conversations search comepleted", true, conversations)
}

func PinConversation(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in conversation member id"), StatusCode: http.StatusBadRequest})
		return
	}
	pinned := convUseCases.PinConversationUseCase(ctx, &id)
	if pinned != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "conversation pinned", true, nil)
}

func UnPinConversation(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in conversation member id"), StatusCode: http.StatusBadRequest})
		return
	}
	pinned := convUseCases.UnPinConversationMember(ctx, &id)
	if pinned != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "conversation unpinned", true, nil)
}

func FetchPinnedConversations(ctx *gin.Context) {
	convRepo := repository.GetConversationMemberRepo()
	var page int64 = 1
	var limit int64 = 5
	skip := (page - 1) * limit
	convs, err := convRepo.FindManyStripped(map[string]interface{}{
		"userId": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"pinned": true,
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}, options.Find().SetProjection(
		map[string]interface{}{
			"reciepientId":   1,
			"reciepientName": 1,
			"profileImage":   1,
		},
	))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: app_errors.RequestError{StatusCode: http.StatusInternalServerError,
			Err: errors.New("could not retrieve your conversations at this time")}, StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "conversations retrieved", true, convs)
}

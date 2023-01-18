package controllers

import (
	"errors"
	"net/http"

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

	conversations, err := convRepo.FindManyStripped(map[string]interface{}{
		"workspaceId": func() *primitive.ObjectID {
			if ctx.GetString("Workspace") == "" {
				return nil
			}
			return utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
		}(),
		"userId": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
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
	server_response.Response(ctx, http.StatusOK, "conversations fetched", true, conversations)
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

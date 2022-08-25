package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mize.app/app/conversation/models"
	"mize.app/app/conversation/repository"
	"mize.app/app/conversation/usecases/message"
	"mize.app/app_errors"
	"mize.app/server_response"
	"mize.app/utils"
)

func SendMessage(ctx *gin.Context) {
	var payload models.Message
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the required info"), StatusCode: http.StatusBadRequest})
		return
	}
	err := messages.SendMessageUseCase(ctx, payload, ctx.Query("channel"))
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "sent", true, nil)
}

func FetchMessages(ctx *gin.Context) {
	to := ctx.Query("to")
	r := ctx.Query("replyTo")
	var replyTo *primitive.ObjectID
	if to == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the conversation id"), StatusCode: http.StatusBadRequest})
		return
	}
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil || page == 0 {
		page = 1
	}
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 15
	}
	skip := (page - 1) * limit
	if r == "" {
		replyTo = nil
	} else {
		replyTo = utils.HexToMongoId(ctx, r)
	}
	messageRepository := repository.GetMessageRepo()
	messages, err := messageRepository.FindMany(map[string]interface{}{
		"to":          utils.HexToMongoId(ctx, to),
		"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		"replyTo":     replyTo,
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not fetch messages"), StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "messages fetched", true, messages)
}

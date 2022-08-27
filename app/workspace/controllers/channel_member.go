package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/workspace/repository"
	channelMemberUseCases "mize.app/app/workspace/usecases/channel_member"
	"mize.app/app_errors"
	"mize.app/server_response"
	"mize.app/utils"
)

func CreateChannelMember(ctx *gin.Context) {
	channel_id := ctx.Params.ByName("id")
	if channel_id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the valid json"), StatusCode: http.StatusBadGateway})
		return
	}
	id, err := channelMemberUseCases.CreateChannelMemberUseCase(ctx, channel_id, nil, false)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "joined channel successfully", true, id)
}

func AdminAddUserByUsername(ctx *gin.Context) {
	username := ctx.Query("username")
	channel_id := ctx.Query("id")
	if username == "" || channel_id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a username and channel id"), StatusCode: http.StatusBadGateway})
		return
	}
	success := channelMemberUseCases.AddUserByUsernameUseCase(ctx, username, channel_id)
	if !success {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "added to channel successfully", true, nil)
}

func FetchChannels(ctx *gin.Context) {
	channelsRepo := repository.GetChannelMemberRepo()
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil || page == 0 {
		page = 1
	}
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 15
	}
	skip := (page - 1) * limit
	channels, err := channelsRepo.FindManyStripped(map[string]interface{}{
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		"userId":      *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}, options.Find().SetProjection(
		map[string]interface{}{
			"lastMessage":     1,
			"unreadMessages":  1,
			"channelId":       1,
			"channelName":     1,
			"lastMessageSent": 1,
		},
	))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: app_errors.RequestError{StatusCode: http.StatusInternalServerError,
			Err: errors.New("could not retrieve your channels at this time")}, StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "channels retrieved", true, channels)
}

func FetchChannelMembers(ctx *gin.Context) {
	channel_id := ctx.Query("id")
	if channel_id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in channel id"), StatusCode: http.StatusBadRequest})
		return
	}
	channelMemberRepo := repository.GetChannelMemberRepo()
	exists, err := channelMemberRepo.CountDocs(map[string]interface{}{
		"channelId":   *utils.HexToMongoId(ctx, channel_id),
		"userId":      *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong"), StatusCode: http.StatusInternalServerError})
		return
	}
	if exists == 0 {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you are not a member of this channel"), StatusCode: http.StatusUnauthorized})
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

	members, err := channelMemberRepo.FindManyStripped(map[string]interface{}{
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		"channelId":   *utils.HexToMongoId(ctx, channel_id),
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}, options.Find().SetProjection(
		map[string]interface{}{
			"userName": 1,
			"admin":    1,
		},
	))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not fetch members"), StatusCode: http.StatusInternalServerError})
		return
	}

	server_response.Response(ctx, http.StatusOK, "members fetched", true, members)
}

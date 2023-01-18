package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/workspace"
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

func AdminAddUserToChannel(ctx *gin.Context) {
	username := ctx.Query("username")
	addBy := ctx.Query("type")
	channel_id := ctx.Query("id")
	if username == "" || channel_id == "" || addBy == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a username, type and channel id"), StatusCode: http.StatusBadGateway})
		return
	}
	success := channelMemberUseCases.AddUserByUsernameUseCase(ctx, username, channel_id, addBy)
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
			"profileImage":    1,
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

func LeaveChannel(ctx *gin.Context) {
	channel_id := ctx.Query("id")
	if channel_id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in channel id"), StatusCode: http.StatusBadRequest})
		return
	}
	deleted := channelMemberUseCases.LeaveChannelUseCase(ctx, &channel_id)
	if !deleted {
		return
	}
	server_response.Response(ctx, http.StatusOK, "channel exited", true, nil)
}

func PinChannel(ctx *gin.Context) {
	channel_id := ctx.Query("id")
	if channel_id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in channel id"), StatusCode: http.StatusBadRequest})
		return
	}
	pinned := channelMemberUseCases.PinChannelMember(ctx, &channel_id)
	if !pinned {
		return
	}

	server_response.Response(ctx, http.StatusOK, "channel pinned", true, nil)
}

func UnPinChannel(ctx *gin.Context) {
	channel_id := ctx.Query("id")
	if channel_id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in channel id"), StatusCode: http.StatusBadRequest})
		return
	}
	pinned := channelMemberUseCases.UnPinChannelMember(ctx, &channel_id)
	if !pinned {
		return
	}

	server_response.Response(ctx, http.StatusOK, "channel unpinned", true, nil)
}

func FetchPinnedChannels(ctx *gin.Context) {
	channelsRepo := repository.GetChannelMemberRepo()
	var page int64 = 1
	var limit int64 = 5
	skip := (page - 1) * limit
	channels, err := channelsRepo.FindManyStripped(map[string]interface{}{
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		"userId":      *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"pinned":      true,
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}, options.Find().SetProjection(
		map[string]interface{}{
			"channelId":    1,
			"channelName":  1,
			"profileImage": 1,
		},
	))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: app_errors.RequestError{StatusCode: http.StatusInternalServerError,
			Err: errors.New("could not retrieve your channels at this time")}, StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "channels retrieved", true, channels)
}

func AddChannelAdminAccess(ctx *gin.Context) {
	channel_id := ctx.Params.ByName("id")
	if channel_id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the channel id"), StatusCode: http.StatusBadRequest})
		return
	}
	var payload workspace.AddChannelPiviledges
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the appropraite payload"), StatusCode: http.StatusBadRequest})
		return
	}
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	err := channelMemberUseCases.AddChannelAdminAccessUseCase(ctx, payload.Id, channel_id, payload.Permissions)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "priviledges updated", true, nil)
}

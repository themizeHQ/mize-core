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
	id, err := channelMemberUseCases.CreateChannelMemberUseCase(ctx, channel_id, false)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "joined channel successfully", true, id)
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
	channels, err := channelsRepo.FindMany(map[string]interface{}{
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		"userId":      *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	},
	)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: app_errors.RequestError{StatusCode: http.StatusInternalServerError,
			Err: errors.New("could not retrieve your channels at this time")}, StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "channels retrieved", true, channels)
}

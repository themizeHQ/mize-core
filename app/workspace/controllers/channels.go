package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	convRepo "mize.app/app/conversation/repository"
	"mize.app/app/media"
	"mize.app/app/workspace"
	"mize.app/app/workspace/models"
	"mize.app/app/workspace/repository"
	channelUseCases "mize.app/app/workspace/usecases/channel"
	channelmembersUseCases "mize.app/app/workspace/usecases/channel_member"
	"mize.app/app_errors"
	"mize.app/authentication"
	channelConstants "mize.app/constants/channel"
	mediaConstants "mize.app/constants/media"
	"mize.app/constants/message"
	"mize.app/logger"
	"mize.app/server_response"
	"mize.app/utils"
)

var waitGroup sync.WaitGroup

func CreateChannel(ctx *gin.Context) {
	var payload []models.Channel
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json array value"), StatusCode: http.StatusBadGateway})
		return
	}
	channels, err := channelUseCases.CreateChannelUseCase(ctx, payload)
	if err != nil {
		return
	}
	for _, channel := range channels {
		waitGroup.Add(1)
		go func(id string, name *string) {
			defer func() {
				waitGroup.Done()
			}()
			channelmembersUseCases.CreateChannelMemberUseCase(ctx, id, name, true)
		}(channel.Id.Hex(), &channel.Name)
	}
	waitGroup.Wait()
	server_response.Response(ctx, http.StatusCreated, "channel successfully created.", true, nil)
}

func DeleteChannel(ctx *gin.Context) {
	channel_id := ctx.Params.ByName("id")
	success, err := channelUseCases.DeleteChannelUseCase(ctx, channel_id)
	if err != nil || !success {
		return
	}
	server_response.Response(ctx, http.StatusOK, "channel deleted successfully", success, nil)
}

func UpdateChannelInfo(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the channel id"), StatusCode: http.StatusBadRequest})
		return
	}
	var payload workspace.UpdateChannelType
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the required payload"), StatusCode: http.StatusBadRequest})
		return
	}
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
	}
	err := channelUseCases.UpdateChannelInfoUseCase(ctx, id, payload)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "channel updated", true, nil)
}

func UpdateChannelProfileImage(ctx *gin.Context) {
	channel_id := ctx.Query("id")

	if channel_id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("set channel id to update"), StatusCode: http.StatusBadRequest})
		return
	}
	file, fileHeader, err := ctx.Request.FormFile("media")
	defer func() {
		file.Close()
	}()
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}

	success, err := authentication.HasChannelAccess(ctx, channel_id, []channelConstants.ChannelAdminAccess{channelConstants.CHANNEL_INFO_EDIT_ACCESS})
	if err != nil || !success {
		return
	}

	uploadRepo := media.GetUploadRepo()
	profileImageUpload, err := uploadRepo.FindOneByFilter(map[string]interface{}{
		"uploadBy": *utils.HexToMongoId(ctx, channel_id),
		"type":     mediaConstants.CHANNEL_IMAGE,
	}, options.FindOne().SetProjection(map[string]interface{}{
		"publicId": 1,
	}))
	if err != nil {
		err := errors.New("could not update profile image")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	var data *media.Upload
	if profileImageUpload == nil {
		data, err = media.UploadToCloudinary(ctx, file, "/core/channel-images", nil)
		if err != nil {
			return
		}
	} else {
		data, err = media.UploadToCloudinary(ctx, file, "/core/channel-images", &profileImageUpload.PublicID)
		if err != nil {
			return
		}
	}
	data.Type = mediaConstants.CHANNEL_IMAGE
	data.Format = strings.Split(fileHeader.Filename, ".")[len(strings.Split(fileHeader.Filename, "."))-1]
	data.UploadBy = *utils.HexToMongoId(ctx, channel_id)

	if profileImageUpload == nil {
		err = channelUseCases.UploadChannelImage(ctx, data, channel_id)
		if err != nil {
			return
		}
	} else {
		err = channelUseCases.UpdateChannelImage(ctx, data.Bytes, channel_id)
		if err != nil {
			return
		}
	}
	server_response.Response(ctx, http.StatusCreated, "upload success", true, nil)
}

func FetchChannelMedia(ctx *gin.Context) {
	channel := ctx.Query("channel")
	if channel == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a channel id"), StatusCode: http.StatusBadRequest})
		return
	}
	mediaType := ctx.Query("type")
	mediaTypes := []string{"image", "video", "audio", "docs", "links"}
	match := false
	for _, t := range mediaTypes {
		if t == mediaType {
			match = true
			break
		}
	}
	if !match {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("valid media types are image, video, audio, docs, links"), StatusCode: http.StatusBadRequest})
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
	messageRepo := convRepo.GetMessageRepo()
	var resources *[]map[string]interface{}
	if mediaType == "image" {
		resources, err = messageRepo.FindManyStripped(map[string]interface{}{
			"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			"to":          *utils.HexToMongoId(ctx, channel),
			"type":        message.IMAGE_MESSAGE,
		}, &options.FindOptions{
			Limit: &limit,
			Skip:  &skip,
		}, options.Find().SetProjection(map[string]int{
			"resourceUrl": 1,
			"userName":    1,
		}))
	} else if mediaType == "video" {
		resources, err = messageRepo.FindManyStripped(map[string]interface{}{
			"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			"to":          *utils.HexToMongoId(ctx, channel),
			"type":        message.VIDEO_MESSAGE,
		}, &options.FindOptions{
			Limit: &limit,
			Skip:  &skip,
		}, options.Find().SetProjection(map[string]int{
			"resourceUrl": 1,
			"userName":    1,
		}))
	} else if mediaType == "audio" {
		resources, err = messageRepo.FindManyStripped(map[string]interface{}{
			"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			"to":          *utils.HexToMongoId(ctx, channel),
			"type":        message.AUDIO_MESSAGE,
		}, &options.FindOptions{
			Limit: &limit,
			Skip:  &skip,
		}, options.Find().SetProjection(map[string]int{
			"resourceUrl": 1,
			"userName":    1,
		}))
	}
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not fetch resources"), StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "resources fetched", true, resources)
}

func FetchAllChannels(ctx *gin.Context) {
	admin := ctx.Query("admin")
	channelsRepo := repository.GetChannelRepo()
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
		"private": func() map[string]interface{} {
			if admin == "true" && authentication.RoleType(ctx.GetString("Role")) == authentication.ADMIN {
				return map[string]interface{}{
					"$in": []bool{true, false},
				}
			}
			return map[string]interface{}{
				"$in": []bool{false},
			}
		}(),
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}, options.Find().SetProjection(func() map[string]int {
		if admin == "true" && authentication.RoleType(ctx.GetString("Role")) == authentication.ADMIN {
			return map[string]int{
				"compulsory":   1,
				"private":      1,
				"channelId":    1,
				"name":         1,
				"membersCount": 1,
				"profileImage": 1,
			}
		}
		return map[string]int{
			"channelId":    1,
			"membersCount": 1,
			"name":         1,
			"profileImage": 1,
		}
	}(),
	))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: app_errors.RequestError{StatusCode: http.StatusInternalServerError,
			Err: errors.New("could not retrieve your channels at this time")}, StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "channels retrieved", true, channels)

}

func FetchChannelDetails(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a valid channel id"), StatusCode: http.StatusBadRequest})
		return
	}
	channelRepo := repository.GetChannelRepo()
	channel, err := channelRepo.FindOneByFilter(map[string]interface{}{
		"_id":         id,
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	})
	if err != nil {
		logger.Error(errors.New("fetch workspace details"), zap.Error(err))
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not fetch workspace"), StatusCode: http.StatusInternalServerError})
		return
	}
	if channel == nil {
		server_response.Response(ctx, http.StatusNotFound, "channel not fetched", false, nil)
		return
	}
	server_response.Response(ctx, http.StatusOK, "channel fetched", true, channel)
}

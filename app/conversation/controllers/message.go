package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mize.app/app/conversation/models"
	"mize.app/app/conversation/repository"
	"mize.app/app/conversation/usecases/message"
	"mize.app/app/media"
	channelsRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	mediaConstants "mize.app/constants/media"
	messageConstants "mize.app/constants/message"
	"mize.app/server_response"
	"mize.app/utils"
)

func SendMessage(ctx *gin.Context) {
	raw_form, err := ctx.MultipartForm()
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in message data"), StatusCode: http.StatusBadRequest})
		return
	}
	parsed_form := map[string]interface{}{}
	for key, value := range raw_form.Value {
		parsed_form[key] = value[0]
	}
	message := models.Message{}
	pJson, err := json.Marshal(parsed_form)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	err = json.Unmarshal(pJson, &message)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	file, fileHeader, err := ctx.Request.FormFile("resource")
	if err != nil {
		if err.Error() == "http: no such file" {
			message.Type = messageConstants.TEXT_MESSAGE
		} else {
			err := errors.New("could not process media file")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
			return
		}
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	var data *media.Upload
	if message.Type == messageConstants.MessageType(messageConstants.IMAGE_MESSAGE) {
		data, err = media.UploadToCloudinary(ctx, file, fmt.Sprintf("/core/message-images/%s/%s", ctx.GetString("Workspace"), message.To.Hex()), nil)
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not send image"), StatusCode: http.StatusInternalServerError})
			return
		}
		data.Type = mediaConstants.MESSAGE_IMAGE
	}
	if message.Type == messageConstants.MessageType(messageConstants.AUDIO_MESSAGE) {
		data, err = media.UploadToCloudinaryRAW(ctx, file, fmt.Sprintf("/core/message-audio/%s/%s", ctx.GetString("Workspace"), message.To.Hex()), nil, "mp3")
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not send audio"), StatusCode: http.StatusInternalServerError})
			return
		}
		data.Type = mediaConstants.MESSAGE_AUDIO
	}
	if message.Type == messageConstants.MessageType(messageConstants.VIDEO_MESSAGE) {
		data, err = media.UploadToCloudinaryRAW(ctx, file, fmt.Sprintf("/core/message-video/%s/%s", ctx.GetString("Workspace"), message.To.Hex()), nil, "mp4")
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not send video"), StatusCode: http.StatusInternalServerError})
			return
		}
		data.Type = mediaConstants.MESSAGE_VIDEO
	}
	if message.Type != messageConstants.MessageType(messageConstants.TEXT_MESSAGE) {
		data.Format = strings.Split(fileHeader.Filename, ".")[len(strings.Split(fileHeader.Filename, "."))-1]
		data.UploadBy = *utils.HexToMongoId(ctx, ctx.GetString("UserId"))
		data.FileName = fileHeader.Filename
		message.ResourceUrl = &data.Url
	}

	err = messages.SendMessageUseCase(ctx, message, ctx.Query("channel"), data)
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
	msgChan := make(chan *[]models.Message)
	go func(msg chan *[]models.Message) {
		m, err := messageRepository.FindMany(map[string]interface{}{
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
		msg <- m
	}(msgChan)
	if page == 1 {
		go func() {
			channelsRepo := channelsRepo.GetChannelMemberRepo()
			channelsRepo.UpdatePartialByFilter(ctx, map[string]interface{}{
				"channelId": utils.HexToMongoId(ctx, to),
				"userId":    utils.HexToMongoId(ctx, ctx.GetString("UserId")),
			}, map[string]interface{}{
				"unreadMessages": 0,
			})
		}()
	}
	server_response.Response(ctx, http.StatusOK, "messages fetched", true, <-msgChan)
}

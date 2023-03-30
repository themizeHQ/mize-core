package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mize.app/app/conversation/models"
	"mize.app/app/conversation/repository"
	types "mize.app/app/conversation/types"
	convUsecases "mize.app/app/conversation/usecases/conversation"
	messages "mize.app/app/conversation/usecases/message"
	"mize.app/app/media"
	userRepository "mize.app/app/user/repository"
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
	new := ctx.Query("new")
	if new == "true" {
		reciepientId := ctx.Query("reciepientId")
		if reciepientId == "" {
			server_response.Response(ctx, http.StatusBadRequest, "pass in a recipientID to be able to create a conversation", false, nil)
			return
		}
		if reciepientId == ctx.GetString("UserId") {
			server_response.Response(ctx, http.StatusBadRequest, "you cannot send a message to yourself", false, nil)
			return
		}
		convRepo := repository.GetConversationMemberRepo()
		exists, err := convRepo.CountDocs(map[string]interface{}{
			"participants": map[string]interface{}{
				"$in": []interface{}{reciepientId, *utils.HexToMongoId(ctx, ctx.GetString("UserId"))},
			},
		})
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not create conversation"), StatusCode: http.StatusBadRequest})
			return
		}
		if exists == 0 {
			convId, reciepientName, recipientImage, recipientImageThumbnail, err := convUsecases.CreateConversationUseCase(ctx, reciepientId)
			if err != nil {
				return
			}
			var wg sync.WaitGroup
			chan1 := make(chan error)
			wg.Add(1)
			go func(e chan error) {
				defer func() {
					wg.Done()
				}()
				_, err := convUsecases.CreateConversationMemberUseCase(ctx, convId, *reciepientName, ctx.GetString("UserId"), recipientImage, recipientImageThumbnail)
				e <- err
			}(chan1)
			wg.Add(1)
			chan2 := make(chan error)
			go func(e chan error) {
				defer func() {
					wg.Done()
				}()
				userRepo := userRepository.GetUserRepo()
				user, err := userRepo.FindById(ctx.GetString("UserId"), options.FindOne().SetProjection(map[string]interface{}{
					"profileImage":          1,
					"profileImageThumbnail": 1,
				}))
				if err != nil {
					e <- err
					return
				}
				_, err = convUsecases.CreateConversationMemberUseCase(ctx, convId, ctx.GetString("Username"), reciepientId, user.ProfileImage, user.ProfileImageThumbNail)
				e <- err
			}(chan2)

			err1 := <-chan1
			err2 := <-chan2
			if err1 != nil || err2 != nil {
				server_response.Response(ctx, http.StatusInternalServerError, "could not create conversation members", true, nil)
				return
			}
			wg.Wait()
			message.To = *convId
		}
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
	message.From = *utils.HexToMongoId(ctx, ctx.GetString("UserId"))
	message.ProfileImage = ctx.GetString("ProfileImage")
	message.Username = ctx.GetString("Username")

	err = messages.SendMessageUseCase(ctx, message, ctx.Query("channel"), data)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "sent", true, nil)
}

func FetchMessages(ctx *gin.Context) {
	to := ctx.Query("to")
	r := ctx.Query("replyTo")
	new := ctx.Query("new")
	var replyTo *primitive.ObjectID
	if to == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the conversation id"), StatusCode: http.StatusBadRequest})
		return
	}
	startID := ctx.Query("id")
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 15
	}
	if r == "" {
		replyTo = nil
	} else {
		replyTo = utils.HexToMongoId(ctx, r)
	}
	messageRepository := repository.GetMessageRepo()
	msgChan := make(chan *[]models.Message)
	go func(msg chan *[]models.Message) {
		m, err := messageRepository.FindMany(map[string]interface{}{
			"to":      utils.HexToMongoId(ctx, to),
			"replyTo": replyTo,
			"_id": func() map[string]interface{} {
				if startID == "" {
					return map[string]interface{}{"$gt": primitive.NilObjectID}
				}
				if new == "true" {
					return map[string]interface{}{"$gt": *utils.HexToMongoId(ctx, startID)}
				}
				return map[string]interface{}{"$lt": *utils.HexToMongoId(ctx, startID)}
			}(),
		}, options.Find().SetSort(map[string]interface{}{
			"createdAt": -1,
		}), options.Find().SetSort(map[string]interface{}{
			"_id": -1,
		}), &options.FindOptions{
			Limit: &limit,
		})
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not fetch messages"), StatusCode: http.StatusBadRequest})
			return
		}
		msg <- m
	}(msgChan)
	if new == "true" || startID == "" {
		go func() {
			channelMemberRepo := channelsRepo.GetChannelMemberRepo()
			channelMemberRepo.UpdatePartialByFilter(ctx, map[string]interface{}{
				"channelId": utils.HexToMongoId(ctx, to),
				"userId":    utils.HexToMongoId(ctx, ctx.GetString("UserId")),
			}, map[string]interface{}{
				"unreadMessages": 0,
			})
		}()
		go func() {
			convMemberRepo := repository.GetConversationMemberRepo()
			convMemberRepo.UpdatePartialByFilter(ctx, map[string]interface{}{
				"conversationId": utils.HexToMongoId(ctx, to),
				"userId":         utils.HexToMongoId(ctx, ctx.GetString("UserId")),
			}, map[string]interface{}{
				"unreadMessages": 0,
			})
		}()
	}
	server_response.Response(ctx, http.StatusOK, "messages fetched", true, <-msgChan)
}

func DeleteMessages(ctx *gin.Context) {
	var payload []types.DeleteMessage
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("provide channels and messages ids"), StatusCode: http.StatusBadRequest})
		return
	}
	err := messages.DeleteMessageUseCase(ctx, payload)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "messages deleted", true, nil)
}

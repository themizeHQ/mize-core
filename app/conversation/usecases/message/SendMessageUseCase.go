package messages

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"mize.app/app/conversation/models"
	conversationRepository "mize.app/app/conversation/repository"
	"mize.app/app/media"
	channelRepository "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	// mediaConstants "mize.app/constants/media"
	"mize.app/emitter"
	"mize.app/utils"
)

func SendMessageUseCase(ctx *gin.Context, payload models.Message, channel string, upload *media.Upload) error {
	messageRepository := conversationRepository.GetMessageRepo()
	channelRepository := channelRepository.GetChannelMemberRepo()
	if channel == "true" {
		exist, err := channelRepository.CountDocs(map[string]interface{}{
			"channelId": utils.HexToMongoId(ctx, payload.To.Hex()),
			"userId":    utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		})
		if err != nil {
			err = errors.New("an error occured")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return err
		}
		if exist != 1 {
			err = errors.New("you are not a member of this channel")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
			return err
		}
	}
	var replyTo string
	if payload.ReplyTo != nil {
		exist, err := messageRepository.CountDocs(map[string]interface{}{
			"_id": payload.ReplyTo.Hex(),
		})
		if err != nil {
			err = errors.New("an error occured")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return err
		}
		if exist != 1 {
			err = errors.New("message has been deleted")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
			return err
		}
		replyTo = payload.ReplyTo.Hex()
	}
	payload.WorkspaceId = *utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
	payload.From = *utils.HexToMongoId(ctx, ctx.GetString("UserId"))
	payload.Username = ctx.GetString("Username")
	_, err := messageRepository.CreateOne(payload)
	if err != nil {
		fmt.Println(err)
		err = errors.New("message could not be sent")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return err
	}
	var wg sync.WaitGroup
	if channel == "true" {
		chan1 := make(chan error)
		err := messageRepository.StartTransaction(ctx, func(sc *mongo.SessionContext, c *context.Context) error {
			wg.Add(1)
			go func(ch chan error) {
				defer func() {
					wg.Done()
				}()
				success, err := channelRepository.UpdatePartialByFilter(ctx, map[string]interface{}{
					"channelId": utils.HexToMongoId(ctx, payload.To.Hex()),
					"userId":    utils.HexToMongoId(ctx, ctx.GetString("UserId")),
				}, map[string]interface{}{
					"lastMessage":     payload.Text,
					"lastMessageSent": primitive.NewDateTimeFromTime(time.Now()),
				})
				if err != nil || !success {
					ch <- err
					(*sc).AbortTransaction(*c)
				}
				ch <- nil
			}(chan1)

			chan2 := make(chan error)
			wg.Add(1)
			go func(ch chan error) {
				defer func() {
					wg.Done()
				}()
				success, err := channelRepository.UpdateWithOperator(map[string]interface{}{
					"channelId": utils.HexToMongoId(ctx, payload.To.Hex()),
					"userId":    utils.HexToMongoId(ctx, ctx.GetString("UserId")),
				}, map[string]interface{}{
					"$inc": map[string]interface{}{
						"unreadMessages": 1,
					}},
				)
				if err != nil || !success {
					ch <- err
					(*sc).AbortTransaction(*c)
				}
				ch <- nil
			}(chan2)

			chan3 := make(chan error)
			wg.Add(1)
			go func(ch chan error) {
				defer func() {
					wg.Done()
				}()
				uploadRepo := media.GetUploadRepo()
				_, err := uploadRepo.CreateOne(*upload)
				if err != nil {
					ch <- err
					(*sc).AbortTransaction(*c)
				}
				ch <- nil
			}(chan3)

			err1 := <-chan1
			err2 := <-chan2
			err3 := <-chan3
			if err1 != nil || err2 != nil || err3 != nil {
				return errors.New("message could not be sent")
			}
			wg.Wait()
			return nil
		})
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return err
		}
	}
	emitter.Emitter.Emit(emitter.Events.MESSAGES_EVENTS.MESSAGE_SENT, map[string]interface{}{
		"time":        time.Now(),
		"from":        payload.From.Hex(),
		"to":          payload.To.Hex(),
		"text":        payload.Text,
		"replyTo":     replyTo,
		"username":    payload.Username,
		"resourceUrl": payload.ResourceUrl,
		"type":        payload.Type,
	})
	return nil
}

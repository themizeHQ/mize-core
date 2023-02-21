package messages

import (
	"context"
	"errors"

	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mize.app/app/conversation/models"
	conversationRepository "mize.app/app/conversation/repository"
	"mize.app/app/media"
	channelRepository "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/emitter"
	"mize.app/realtime"
	"mize.app/utils"
)

func SendMessageUseCase(ctx *gin.Context, payload models.Message, channel string, upload *media.Upload) error {
	messageRepository := conversationRepository.GetMessageRepo()
	convMemberRepository := conversationRepository.GetConversationMemberRepo()
	channelMemberRepository := channelRepository.GetChannelMemberRepo()
	var wg sync.WaitGroup
	senderImgURLChan := make(chan *string)
	channelMemberID := make(chan string)
	var senderImgURL *string
	if channel == "true" {
		payload.WorkspaceId = utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
		chan1 := make(chan error)
		wg.Add(1)
		go func(e chan error, img chan *string, cmID chan string) {
			defer func() {
				wg.Done()
			}()

			exist, err := channelMemberRepository.FindOneByFilter(map[string]interface{}{
				"channelId": *utils.HexToMongoId(ctx, payload.To.Hex()),
				"userId":    *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
			}, options.FindOne().SetProjection(map[string]interface{}{
				"profileImage": 1,
			}))
			if err != nil {
				e <- errors.New("an error occured")
				img <- nil
				channelMemberID <- ""
				return
			}
			if exist == nil {
				e <- errors.New("you are not a member of this channel")
				img <- nil
				channelMemberID <- ""
				return
			}
			e <- nil
			img <- exist.ProfileImage
			channelMemberID <- exist.Id.Hex()
		}(chan1, senderImgURLChan, channelMemberID)
		err1 := <-chan1
		if err1 != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err1, StatusCode: http.StatusInternalServerError})
			return err1
		}
		senderImgURL = <-senderImgURLChan
		payload.From = *utils.HexToMongoId(ctx, <-channelMemberID)
		payload.WorkspaceId = utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
	} else {
		chan1 := make(chan error)
		convMemberID := make(chan string)
		wg.Add(1)
		go func(e chan error, cmID chan string) {
			defer func() {
				wg.Done()
			}()

			exist, err := convMemberRepository.FindOneByFilter(map[string]interface{}{
				"userId":         *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
				"conversationId": payload.To,
			}, options.FindOne().SetProjection(map[string]interface{}{
				"_id": 1,
			}))
			if err != nil {
				e <- errors.New("an error occured")
				cmID <- ""
				return
			}
			if exist == nil {
				e <- errors.New("invalid conversation id provided")
				cmID <- ""
				return
			}
			e <- nil
			cmID <- exist.Id.Hex()
		}(chan1, convMemberID)
		err1 := <-chan1
		payload.From = *utils.HexToMongoId(ctx, <-convMemberID)
		if err1 != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err1, StatusCode: http.StatusInternalServerError})
			return err1
		}
	}

	var replyTo string
	if payload.ReplyTo != nil {
		chan1 := make(chan error)
		wg.Add(1)
		go func(e chan error) {
			defer func() {
				wg.Done()
			}()

			exist, err := messageRepository.CountDocs(map[string]interface{}{
				"_id": payload.ReplyTo.Hex(),
			})
			if err != nil {
				e <- errors.New("an error occured")
				return
			}
			if exist != 1 {
				e <- errors.New("message has been deleted")
				return
			}
			e <- nil
			replyTo = payload.ReplyTo.Hex()
		}(chan1)
		err1 := <-chan1
		if err1 != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err1, StatusCode: http.StatusInternalServerError})
			return err1
		}
	}
	if channel == "true" {
		chan1 := make(chan error)
		err := messageRepository.StartTransaction(ctx, func(sc *mongo.SessionContext, c *context.Context) error {
			wg.Add(1)
			go func(ch chan error) {
				defer func() {
					wg.Done()
				}()
				success, err := channelMemberRepository.UpdatePartialByFilter(ctx, map[string]interface{}{
					"channelId": utils.HexToMongoId(ctx, payload.To.Hex()),
				}, map[string]interface{}{
					"lastMessage":     payload.Text,
					"lastMessageSent": primitive.NewDateTimeFromTime(time.Now()),
				})
				if err != nil || !success {
					ch <- err
					(*sc).AbortTransaction(*c)
				}
				ch <- nil
				(*sc).CommitTransaction(*c)
			}(chan1)

			chan2 := make(chan error)
			wg.Add(1)
			go func(ch chan error) {
				defer func() {
					wg.Done()
				}()
				_, err := messageRepository.CreateOne(payload)
				if err != nil {
					err = errors.New("message could not be sent")
					app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
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
				success, err := channelMemberRepository.UpdateWithOperator(map[string]interface{}{
					"channelId": utils.HexToMongoId(ctx, payload.To.Hex()),
					"userId": map[string]interface{}{
						"$ne": utils.HexToMongoId(ctx, ctx.GetString("UserId")),
					},
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
			}(chan3)

			var err4 error
			if upload != nil {
				chan4 := make(chan error)
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
				}(chan4)
				err4 = <-chan4
			}

			err1 := <-chan1
			err2 := <-chan2
			err3 := <-chan3
			if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
				return errors.New("message could not be sent")
			}
			return nil
		})
		wg.Wait()
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return err
		}
	} else {
		err := messageRepository.StartTransaction(ctx, func(sc *mongo.SessionContext, c *context.Context) error {
			chan1 := make(chan error)
			wg.Add(1)
			go func(ch chan error) {
				defer func() {
					wg.Done()
				}()
				success, err := convMemberRepository.UpdatePartialByFilter(ctx, map[string]interface{}{
					"conversationId": utils.HexToMongoId(ctx, payload.To.Hex()),
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
				_, err := messageRepository.CreateOne(payload)
				if err != nil {
					err = errors.New("message could not be sent")
					app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
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
				success, err := convMemberRepository.UpdateManyWithOperator(map[string]interface{}{
					"conversationId": utils.HexToMongoId(ctx, payload.To.Hex()),
					"_id": map[string]interface{}{
						"$ne": payload.From,
					},
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
			}(chan3)

			var err4 error
			if upload != nil {
				chan4 := make(chan error)
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
				}(chan4)
				err4 = <-chan4
			}

			err1 := <-chan1
			err2 := <-chan2
			err3 := <-chan3
			if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
				return errors.New("message could not be sent")
			}
			return nil
		})
		wg.Wait()
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return err
		}
	}
	realtime.CentrifugoController.Publish(payload.To.Hex(), realtime.MessageScope.CONVERSATION, map[string]interface{}{
		"time":        time.Now(),
		"from":        payload.From.Hex(),
		"to":          payload.To.Hex(),
		"text":        payload.Text,
		"replyTo":     replyTo,
		"username":    payload.Username,
		"resourceUrl": payload.ResourceUrl,
		"type":        payload.Type,
	})
	if channel == "true" {
		emitter.Emitter.Emit(emitter.Events.MESSAGES_EVENTS.MESSAGE_SENT, map[string]interface{}{
			"channel":      payload.To.Hex(),
			"by":           payload.Username,
			"msg":          payload.Text,
			"resourceID":   payload.Id,
			"profileImage": senderImgURL,
		})
	}
	return nil
}

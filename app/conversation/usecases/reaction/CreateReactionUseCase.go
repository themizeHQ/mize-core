package reaction

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"mize.app/app/conversation/repository"
	"mize.app/app/conversation/types"
	channelRepository "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/logger"
	"mize.app/utils"
)

func CreateReactionUseCase(ctx *gin.Context, data types.Reaction, channel bool) error {
	if !utils.IsEmoji(data.Reaction) {
		err := errors.New("reaction must be an emoji")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}

	convMemberRepository := repository.GetConversationMemberRepo()
	channelRepository := channelRepository.GetChannelMemberRepo()
	var wg sync.WaitGroup
	var convId = make(chan string)
	if channel {
		chan1 := make(chan error)
		wg.Add(1)
		go func(e chan error, cID chan string) {
			defer func() {
				wg.Done()
			}()

			exist, err := channelRepository.FindOneByFilter(map[string]interface{}{
				"channelId": data.ConversationID,
				"userId":    *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
			}, options.FindOne().SetProjection(map[string]interface{}{
				"_id": 1,
			}))
			if err != nil {
				e <- errors.New("an error occured")
				cID <- ""
				return
			}
			if exist == nil {
				e <- errors.New("you are not a member of this channel")
				cID <- ""
				return
			}
			e <- nil
			cID <- exist.Id.Hex()
		}(chan1, convId)
		err1 := <-chan1
		if err1 != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err1, StatusCode: http.StatusInternalServerError})
			return err1
		}
	} else {
		chan1 := make(chan error)
		wg.Add(1)
		go func(e chan error, cID chan string) {
			defer func() {
				wg.Done()
			}()

			exist, err := convMemberRepository.FindOneByFilter(map[string]interface{}{
				// "userId":         *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
				"conversationId": data.ConversationID,
			}, options.FindOne().SetProjection(map[string]interface{}{
				"_id": 1,
			}))

			if err != nil {
				e <- errors.New("an error occured")
				cID <- ""
				return
			}
			if exist == nil {
				e <- errors.New("invalid conversation id provided")
				cID <- ""
				return
			}
			e <- nil
			cID <- exist.Id.Hex()
		}(chan1, convId)
		err1 := <-chan1
		if err1 != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err1, StatusCode: http.StatusInternalServerError})
			return err1
		}
	}
	// convMemberID := <-convId

	reactionRepo := repository.GetReactionRepo()
	return reactionRepo.StartTransaction(ctx, func(sc *mongo.SessionContext, c *context.Context) error {
		count, err := reactionRepo.CountDocs(map[string]interface{}{
			"messageId": data.MessageID,
			"userId":    *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		})
		if err != nil {
			(*sc).AbortTransaction(*c)
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return err
		}
		if count == 0 {
			msgRepo := repository.GetMessageRepo()
			success, err := msgRepo.UpdateWithOperator(map[string]interface{}{
				"_id": data.MessageID,
				"to":  data.ConversationID,
			}, map[string]interface{}{
				"$inc": map[string]interface{}{
					"reactionsCount": 1,
				}})
			if err != nil {
				(*sc).AbortTransaction(*c)
				logger.Error(errors.New("could not update message reaction count"), zap.Error(err))
				app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
				return err
			}
			if !success {
				(*sc).AbortTransaction(*c)
				err = errors.New("could not update message reaction count")
				logger.Error(err)
				app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
				return err
			}
		}
		fmt.Println(map[string]interface{}{
			// "messageId":      data.MessageID,
			// "userId":         convMemberID,
			// "conversationId": data.ConversationID,
			// "workspaceId":    *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			// "userName": ctx.GetString("Username"),
			// "reaction": data.Reaction,
		})
		// success, err := reactionRepo.UpdateOrCreateByField(map[string]interface{}{
		// 	"messageId": data.MessageID,
		// 	"userId":    convMemberID,
		// }, map[string]interface{}{
		// 	"messageId":      data.MessageID,
		// 	"userId":         convMemberID,
		// 	"conversationId": data.ConversationID,
		// 	"workspaceId":    *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		// 	"userName":       ctx.GetString("Username"),
		// 	"reaction":       data.Reaction,
		// }, options.Update().SetUpsert(true))
		// if err != nil {
		// 	fmt.Println("the error", err)
		// 	(*sc).AbortTransaction(*c)
		// 	logger.Error(errors.New("could not update message reaction count"), zap.Error(err))
		// 	app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		// 	return err
		// }
		// if !success {
		// 	(*sc).AbortTransaction(*c)
		// 	err = errors.New("could not update message reaction count")
		// 	logger.Error(err)
		// 	app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		// 	return err
		// }
		return (*sc).CommitTransaction(*c)
	})
}

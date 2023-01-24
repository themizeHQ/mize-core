package reaction

import (
	"context"
	"errors"
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

func RemoveReactionUseCase(ctx *gin.Context, data types.RemoveReaction, channel bool) error {
	convMemberRepository := repository.GetConversationMemberRepo()
	channelRepository := channelRepository.GetChannelMemberRepo()
	var wg sync.WaitGroup
	if channel {
		chan1 := make(chan error)
		wg.Add(1)
		go func(e chan error) {
			defer func() {
				wg.Done()
			}()

			exist, err := channelRepository.FindOneByFilter(map[string]interface{}{
				"channelId": data.ConversationID,
				"userId":    *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
			}, options.FindOne().SetProjection(map[string]interface{}{
				"profileImage": 1,
			}))
			if err != nil {
				e <- errors.New("an error occured")
				return
			}
			if exist == nil {
				e <- errors.New("you are not a member of this channel")
				return
			}
			e <- nil
		}(chan1)
		err1 := <-chan1
		if err1 != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err1, StatusCode: http.StatusInternalServerError})
			return err1
		}
	} else {
		chan1 := make(chan error)
		wg.Add(1)
		go func(e chan error) {
			defer func() {
				wg.Done()
			}()

			exist, err := convMemberRepository.CountDocs(map[string]interface{}{
				"userId":         *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
				"conversationId": data.ConversationID,
			})
			if err != nil {
				e <- errors.New("an error occured")
				return
			}
			if exist != 1 {
				e <- errors.New("invalid conversation id provided")
				return
			}
			e <- nil
		}(chan1)
		err1 := <-chan1
		if err1 != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err1, StatusCode: http.StatusInternalServerError})
			return err1
		}
	}

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
		if count == 1 {
			msgRepo := repository.GetMessageRepo()
			success, err := msgRepo.UpdateWithOperator(map[string]interface{}{
				"_id": data.MessageID,
				"to":  data.ConversationID,
			}, map[string]interface{}{
				"$inc": map[string]interface{}{
					"reactionsCount": -1,
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
		} else {
			return nil
		}
		success, err := reactionRepo.DeleteOne(map[string]interface{}{
			"messageId": data.MessageID,
			"userId":    *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		})
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
		return (*sc).CommitTransaction(*c)
	})
}

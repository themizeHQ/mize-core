package channel

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"mize.app/app/media"
	repository "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func UploadChannelImage(ctx *gin.Context, upload *media.Upload, channel_id string) error {
	uploadRepo := media.GetUploadRepo()
	channelRepo := repository.GetChannelRepo()
	channelMemberRepo := repository.GetChannelMemberRepo()
	err := uploadRepo.StartTransaction(ctx, func(sc *mongo.SessionContext, c *context.Context) error {
		_, err := uploadRepo.CreateOne(*upload)
		if err != nil {
			(*sc).AbortTransaction(*c)
			return err
		}

		_, err = channelRepo.UpdatePartialByFilter(ctx, map[string]interface{}{
			"_id": *utils.HexToMongoId(ctx, channel_id),
		}, map[string]interface{}{
			"profileImage": upload.Url,
		})
		if err != nil {
			(*sc).AbortTransaction(*c)
			return err
		}

		_, err = channelMemberRepo.UpdatePartialByFilter(ctx, map[string]interface{}{
			"channelId": *utils.HexToMongoId(ctx, channel_id),
		}, map[string]interface{}{
			"profileImage": upload.Url,
		})
		if err != nil {
			(*sc).AbortTransaction(*c)
			return err
		}
		(*sc).CommitTransaction(*c)
		return nil
	})

	if err != nil {
		err := errors.New("could not update profile image")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	return err
}

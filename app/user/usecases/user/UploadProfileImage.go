package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"mize.app/app/media"
	userRepository "mize.app/app/user/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func UploadProfileImage(ctx *gin.Context, upload *media.Upload) error {
	uploadRepo := media.GetUploadRepo()
	userRepo := userRepository.GetUserRepo()
	err := uploadRepo.StartTransaction(ctx, func(sc *mongo.SessionContext, c *context.Context) error {
		_, err := uploadRepo.CreateOne(*upload)
		if err != nil {
			(*sc).AbortTransaction(*c)
			return err
		}

		_, err = userRepo.UpdatePartialByFilter(ctx, map[string]interface{}{
			"_id": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
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

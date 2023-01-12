package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"mize.app/app/media"
	"mize.app/app_errors"
)

func UploadProfileImage(ctx *gin.Context, upload *media.Upload) error {
	uploadRepo := media.GetUploadRepo()
	err := uploadRepo.StartTransaction(ctx, func(sc *mongo.SessionContext, c *context.Context) error {
		_, err := uploadRepo.CreateOne(*upload)
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

package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/media"
	"mize.app/app_errors"
)

func UploadProfileImage(ctx *gin.Context, upload *media.Upload) error {
	uploadRepo := media.GetUploadRepo()
	_, err := uploadRepo.CreateOne(*upload)
	if err != nil {
		err := errors.New("could not update profile image")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	return err
}

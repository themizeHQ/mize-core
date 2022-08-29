package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/media"
	"mize.app/app_errors"
	mediaConstants "mize.app/constants/media"
	"mize.app/utils"
)

func UpdateProfileImage(ctx *gin.Context, bytes int) error {
	uploadRepo := media.GetUploadRepo()
	_, err := uploadRepo.UpdatePartialByFilter(ctx, map[string]interface{}{
		"uploadBy": utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"type":     mediaConstants.PROFILE_IMAGE,
	}, map[string]int{
		"bytes": bytes,
	})
	if err != nil {
		err := errors.New("could not update profile image")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	return err
}

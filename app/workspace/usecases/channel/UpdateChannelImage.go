package channel

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/media"
	"mize.app/app_errors"
	"mize.app/utils"
)

func UpdateChannelImage(ctx *gin.Context, bytes int, channel_id string) error {
	uploadRepo := media.GetUploadRepo()
	_, err := uploadRepo.UpdatePartialByFilter(ctx, map[string]interface{}{
		"uploadBy": utils.HexToMongoId(ctx, channel_id),
	}, map[string]int{
		"bytes": bytes,
	})
	if err != nil {
		err := errors.New("could not update channel image")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	return err
}

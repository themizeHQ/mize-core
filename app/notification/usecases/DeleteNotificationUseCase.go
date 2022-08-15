package usecases

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"mize.app/app/notification/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func DeleteNotificationsUseCase(ctx *gin.Context, ids []string) bool {
	notificationRepo := repository.GetNotificationRepo()
	var parsedIds []primitive.ObjectID
	for _, id := range ids {
		parsed := utils.HexToMongoId(ctx, id)
		if parsed == nil {
			return false
		}
		parsedIds = append(parsedIds, *parsed)
	}
	success, err := notificationRepo.DeleteMany(ctx, map[string]interface{}{
		"_id": map[string]interface{}{
			"$in": parsedIds,
		},
		"userId": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return false
	}
	return success
}

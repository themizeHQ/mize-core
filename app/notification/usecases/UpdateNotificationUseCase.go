package usecases

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mize.app/app/notification/repository"
	constnotif "mize.app/constants/notification"
	"mize.app/logger"
	"mize.app/utils"
)

func UpdateNotificationUseCase(ctx *gin.Context, update map[string]interface{}) error {
	notificationRepo := repository.GetNotificationRepo()
	_, err := notificationRepo.UpdatePartialByFilter(ctx, map[string]interface{}{
		"userId": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"scope":  constnotif.USER_NOTIFICATION,
	}, update)
	if err != nil {
		logger.Error(errors.New("failed to update notification"), zap.Error(err))
	}
	return err
}

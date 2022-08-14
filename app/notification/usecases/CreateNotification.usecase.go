package notification

import (
	"github.com/gin-gonic/gin"
	notificationModel "mize.app/app/notification/models"
	notificationRepo "mize.app/app/notification/repository"
)

func CreateNotificationUseCase(ctx *gin.Context, data notificationModel.Notification) {
	notificationRepo := notificationRepo.GetNotificationRepo()
	notificationRepo.CreateOne(data)
}

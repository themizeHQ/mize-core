package usecases

import (
	"github.com/gin-gonic/gin"
	"mize.app/app/notification/models"
	"mize.app/app/notification/repository"
)

func CreateNotificationUseCase(ctx *gin.Context, data models.Notification) {
	notificationRepo := repository.GetNotificationRepo()
	notificationRepo.CreateOne(data)
}

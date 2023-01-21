package usecases

import (
	"github.com/gin-gonic/gin"
	"mize.app/app/notification/models"
	"mize.app/app/notification/repository"
	"mize.app/realtime"
)

func CreateNotificationUseCase(ctx *gin.Context, data models.Notification) {
	notificationRepo := repository.GetNotificationRepo()
	notification, _ := notificationRepo.CreateOne(data)
	realtime.CentrifugoController.Publish(notification.UserId.Hex(), realtime.MessageScope.NOTIFICATION, map[string]interface{}{
		"time":       notification.CreatedAt,
		"resourceId": notification.ResourceId,
		"type":       notification.Type,
		"message":    notification.Message,
		"image":      notification.ImageURL,
		"header":     notification.Header,
		"reacted":    notification.Reacted,
	})
}

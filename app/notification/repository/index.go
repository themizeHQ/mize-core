package notification

import (
	notification "mize.app/app/notification/models"
	db "mize.app/db/mongo"
	repository "mize.app/repository/database/mongo"
)

var NotificationRepository repository.MongoRepository[notification.Notification]

func GetNotificationRepo() repository.MongoRepository[notification.Notification] {
	NotificationRepository = repository.MongoRepository[notification.Notification]{Model: db.Notification}
	return NotificationRepository
}

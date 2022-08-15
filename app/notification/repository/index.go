package repository

import (
	"mize.app/app/notification/models"
	dbMongo "mize.app/db/mongo"
	mongoRepo "mize.app/repository/database/mongo"
)

var NotificationRepository mongoRepo.MongoRepository[models.Notification]

func GetNotificationRepo() mongoRepo.MongoRepository[models.Notification] {
	NotificationRepository = mongoRepo.MongoRepository[models.Notification]{Model: dbMongo.Notification}
	return NotificationRepository
}

package repository

import (
	"mize.app/app/notification/models"
	dbMongo "mize.app/db/mongo"
	mongoRepo "mize.app/repository/database/mongo"
)

var NotificationRepository mongoRepo.MongoRepository[models.Notification]

var AlertRepository mongoRepo.MongoRepository[models.Alert]

func GetNotificationRepo() mongoRepo.MongoRepository[models.Notification] {
	NotificationRepository = mongoRepo.MongoRepository[models.Notification]{Model: dbMongo.Notification}
	return NotificationRepository
}

func GetAlertRepo() mongoRepo.MongoRepository[models.Alert] {
	AlertRepository = mongoRepo.MongoRepository[models.Alert]{Model: dbMongo.Alert}
	return AlertRepository
}

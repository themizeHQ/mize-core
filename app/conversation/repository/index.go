package repository

import (
	"mize.app/app/conversation/models"
	dbMongo "mize.app/db/mongo"
	mongoRepo "mize.app/repository/database/mongo"
)

var MessageRepository mongoRepo.MongoRepository[models.Message]

func GetMessageRepo() mongoRepo.MongoRepository[models.Message] {
	MessageRepository = mongoRepo.MongoRepository[models.Message]{Model: dbMongo.Message}
	return MessageRepository
}

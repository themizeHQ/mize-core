package repository

import (
	"mize.app/app/calls/models"
	dbMongo "mize.app/db/mongo"
	mongoRepo "mize.app/repository/database/mongo"
)

var CallLogRepository mongoRepo.MongoRepository[models.CallLog]

func GetCallLogRepo() mongoRepo.MongoRepository[models.CallLog] {
	CallLogRepository = mongoRepo.MongoRepository[models.CallLog]{Model: dbMongo.Message}
	return CallLogRepository
}

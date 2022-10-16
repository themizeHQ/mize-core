package repository

import (
	"sync"

	dbMongo "mize.app/db/mongo"
	mongoRepo "mize.app/repository/database/mongo"
	"mize.app/app/schedule/models"
)

var once = sync.Once{}

var ScheduleRepository mongoRepo.MongoRepository[models.Schedule]

func GetUploadRepo() mongoRepo.MongoRepository[models.Schedule] {
	once.Do(func() {
		ScheduleRepository = mongoRepo.MongoRepository[models.Schedule]{Model: dbMongo.Upload}
	})
	return ScheduleRepository
}

package repository

import (
	"sync"

	"mize.app/app/schedule/models"
	dbMongo "mize.app/db/mongo"
	mongoRepo "mize.app/repository/database/mongo"
)

var once = sync.Once{}

var ScheduleRepository mongoRepo.MongoRepository[models.Schedule]

func GetScheduleRepo() mongoRepo.MongoRepository[models.Schedule] {
	once.Do(func() {
		ScheduleRepository = mongoRepo.MongoRepository[models.Schedule]{Model: dbMongo.Schedule}
	})
	return ScheduleRepository
}

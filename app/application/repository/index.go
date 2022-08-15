package repository

import (
	"sync"

	"mize.app/app/application/models"
	dbMongo "mize.app/db/mongo"
	mongoRepo "mize.app/repository/database/mongo"
)

var once = sync.Once{}

var AppRepository mongoRepo.MongoRepository[models.Application]

func GetAppRepo() mongoRepo.MongoRepository[models.Application] {
	once.Do(func() {
		AppRepository = mongoRepo.MongoRepository[models.Application]{Model: dbMongo.AppModel}
	})
	return AppRepository
}

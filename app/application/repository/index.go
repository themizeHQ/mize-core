package user

import (
	"sync"

	appModel "mize.app/app/application/models"
	"mize.app/db/mongo"
	"mize.app/repository/database/mongo"
)

var once = sync.Once{}

var AppRepository repository.MongoRepository[appModel.Application]

func GetAppRepo() repository.MongoRepository[appModel.Application] {
	once.Do(func() {
		AppRepository = repository.MongoRepository[appModel.Application]{Model: db.AppModel}
	})
	return AppRepository
}

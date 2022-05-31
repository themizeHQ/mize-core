package user

import (
	appModel "mize.app/app/application/models"
	"mize.app/db/mongo"
	"mize.app/repository/database/mongo"
)

var AppRepository = repository.MongoRepository[appModel.Application]{Model: &db.AppModel}

package repository

import (
	"sync"

	"mize.app/app/user/models"
	dbMongo "mize.app/db/mongo"
	mongoRepo "mize.app/repository/database/mongo"
)

var once = sync.Once{}

var UserRepository mongoRepo.MongoRepository[models.User]

func GetUserRepo() mongoRepo.MongoRepository[models.User] {
	once.Do(func() {
		UserRepository = mongoRepo.MongoRepository[models.User]{Model: dbMongo.UserModel}
	})
	return UserRepository
}

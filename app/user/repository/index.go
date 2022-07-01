package user

import (
	"sync"

	"mize.app/app/user/models"
	"mize.app/db/mongo"
	"mize.app/repository/database/mongo"
)

var once = sync.Once{}

var UserRepository repository.MongoRepository[user.User]

func GetUserRepo() repository.MongoRepository[user.User] {
	once.Do(func() {
		UserRepository = repository.MongoRepository[user.User]{Model: db.UserModel}
	})
	return UserRepository
}

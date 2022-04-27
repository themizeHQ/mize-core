package user

import (
	"mize.app/app/user/models"
	"mize.app/db/mongo"
	"mize.app/repository/database/mongo"
)

var UserRepository = repository.MongoRepository[user.User]{Model: &db.UserModel}

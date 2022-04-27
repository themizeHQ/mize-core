package db

import (
	"context"

	mongoDb "mize.app/db/mongo"
	redisDb "mize.app/db/redis"
)

var mongoContextCancel context.CancelFunc

func ConnectToDb() {
	// connect to mongodb
	mongoContextCancel = mongoDb.ConnectMongo()
	// connect to redis
	redisDb.ConnectRedis()
}

func CleanUp() {
	// cancel the mongodb context
	mongoContextCancel()
}

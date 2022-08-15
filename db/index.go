package db

import (
	"context"

	"mize.app/db/mongo"
	"mize.app/db/redis"
)

var mongoContextCancel context.CancelFunc

func ConnectToDb() {
	// connect to mongodb
	mongoContextCancel = mongo.ConnectMongo()
	// connect to redis
	redis.ConnectRedis()
}

func CleanUp() {
	// cancel the mongodb context
	mongoContextCancel()
}

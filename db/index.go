package db

import (
	"context"

	"mize.app/db/mongo"
	"mize.app/db/redis"
	"mize.app/logger"
)

var mongoContextCancel context.CancelFunc

func ConnectToDb() {
	// connect to mongodb
	mongoContextCancel = mongo.ConnectMongo()
	logger.Info("database (mongodb) - ONLINE")
	// connect to redis
	redis.ConnectRedis()
	logger.Info("database (redis) - ONLINE")
}

func CleanUp() {
	// cancel the mongodb context
	mongoContextCancel()
	logger.Info("resource cleanup (mongodb) - OFFLINE")
}

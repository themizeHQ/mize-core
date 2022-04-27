package db

import (
	mongoDb "mize.app/db/mongo"
	redisDb "mize.app/db/redis"
)

func ConnectToDb() {
	// connect to mongodb
	mongoDb.ConnectMongo()
	// connect to redis
	redisDb.ConnectRedis()
}

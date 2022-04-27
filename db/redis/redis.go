package db

import (
	"os"

	"github.com/go-redis/redis/v8"
)

var (
	Client *redis.Client
)

func ConnectRedis() {
	opt, err := redis.ParseURL(os.Getenv("REDIS_URI"))
	if err != nil {
		panic(err)
	}

	Client = redis.NewClient(opt)
}

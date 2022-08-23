package redis

import (
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

var (
	Client *redis.Client
)

func ConnectRedis() {
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		fmt.Println(err)
		fmt.Println("redis not started")
		return
	}

	Client = redis.NewClient(opt)
}

package repository

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"mize.app/app_errors"
	redisDb "mize.app/db/redis"
)

var (
	RedisRepo RedisRepository
)

type RedisRepository struct {
	Clinet *redis.Client
}

func (redisRepo *RedisRepository) CreateEntry(ctx *gin.Context, key string, payload interface{}, ttl time.Duration) bool {
	c, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer func() {
		cancel()
	}()

	_, err := redisRepo.Clinet.Set(c, key, payload, ttl).Result()

	if err != nil {
		ctx.Abort()
		app_errors.ErrorHandler(ctx, err)
		return false
	}

	return true
}

func SetUpRedisRepo() {
	RedisRepo = RedisRepository{Clinet: redisDb.Client}
}

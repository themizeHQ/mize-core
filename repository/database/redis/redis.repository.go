package repository

import (
	"context"
	"net/http"
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

func generateContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

func (redisRepo *RedisRepository) CreateEntry(ctx *gin.Context, key string, payload interface{}, ttl time.Duration) bool {
	c, cancel := generateContext()

	defer func() {
		cancel()
	}()

	_, err := redisRepo.Clinet.Set(c, key, payload, ttl).Result()

	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return false
	}

	return true
}

func (redisRepo *RedisRepository) FindOne(ctx *gin.Context, key string) *string {
	c, cancel := generateContext()

	defer func() {
		cancel()
	}()

	result, err := redisRepo.Clinet.Get(c, key).Result()

	if err != nil {
		if err.Error() == "redis: nil" {
			return nil
		}
		app_errors.ErrorHandler(ctx, err, http.StatusNotFound)
		return nil
	}
	return &result
}

func (redisRepo *RedisRepository) DeleteOne(ctx *gin.Context, key string) *int {
	c, cancel := generateContext()

	defer func() {
		cancel()
	}()

	result, err := redisRepo.Clinet.Del(c, key).Result()

	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return nil
	}
	result_int := int(result)
	return &result_int
}

func SetUpRedisRepo() {
	RedisRepo = RedisRepository{Clinet: redisDb.Client}
}

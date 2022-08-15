package redis

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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

	return err == nil
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
		return nil
	}
	return &result
}

func (redisRepo *RedisRepository) DeleteOne(ctx *gin.Context, key string) bool {
	c, cancel := generateContext()

	defer func() {
		cancel()
	}()

	result, err := redisRepo.Clinet.Del(c, key).Result()

	if err != nil {
		return false
	}
	if int(result) != 1 {
		return false
	}
	return true
}

func (redisRepo *RedisRepository) CreateInSet(ctx *gin.Context, key string, score float64, member interface{}) bool {
	c, cancel := generateContext()

	defer func() {
		cancel()
	}()

	added := redisRepo.Clinet.ZAdd(c, key, &redis.Z{
		Score: score, Member: member,
	})

	return added != nil
}

func (redisRepo *RedisRepository) FindSet(ctx *gin.Context, key string) *[]string {
	c, cancel := generateContext()

	defer func() {
		cancel()
	}()

	result := redisRepo.Clinet.ZRange(c, key, 0, -1)
	if result == nil {
		return nil
	}
	val := result.Val()
	return &val
}

func SetUpRedisRepo() {
	RedisRepo = RedisRepository{Clinet: redisDb.Client}
}

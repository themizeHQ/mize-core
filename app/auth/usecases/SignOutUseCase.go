package usecases

import (
	"time"

	"github.com/gin-gonic/gin"
	"mize.app/repository/database/redis"
)

func SignOutUseCase(ctx *gin.Context, allDevices string, refreshToken string, accessToken string) {
	if allDevices == "true" {
		redis.RedisRepo.CreateEntry(ctx, ctx.GetString("UserId")+"INVALID_AT", time.Now().Unix(), 10*time.Minute)
	}
	if refreshToken != "" {
		redis.RedisRepo.CreateEntry(ctx, ctx.GetString("UserId")+refreshToken, refreshToken, 2400*time.Hour)
		redis.RedisRepo.CreateEntry(ctx, ctx.GetString("UserId")+accessToken, accessToken, 4*time.Hour)
	}
}

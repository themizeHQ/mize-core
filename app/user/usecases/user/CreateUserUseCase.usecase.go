package user

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	user "mize.app/app/user/models"
	userRepo "mize.app/app/user/repository"
	redis "mize.app/repository/database/redis"
)

func CreateUserUseCase(ctx *gin.Context, userEmail string) (*string, error) {
	var payload user.User
	json.Unmarshal([]byte(*redis.RedisRepo.FindOne(ctx, fmt.Sprintf("%s-user", userEmail))), &payload)
	response := userRepo.UserRepository.CreateOne(ctx, &payload)
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-user", userEmail))
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-otp", userEmail))
	return response, nil
}

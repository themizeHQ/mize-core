package user

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	user "mize.app/app/user/models"
	userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
	redis "mize.app/repository/database/redis"
)

func CreateUserUseCase(ctx *gin.Context, userEmail string) (*mongo.InsertOneResult, error) {
	var payload user.User
	json.Unmarshal([]byte(*redis.RedisRepo.FindOne(ctx, fmt.Sprintf("%s-user", userEmail))), &payload)
	payload.Verified = false
	payload.OrgsCreated = []primitive.ObjectID{}
	err := payload.Validate()
	if err != nil {
		app_errors.ErrorHandler(ctx, err)
		return nil, err
	}
	response := userRepo.UserRepository.CreateOne(ctx, &payload)
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-user", userEmail))
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-otp", userEmail))
	return &response, nil
}

package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/user/models"
	"mize.app/app/user/repository"
	"mize.app/app_errors"
	"mize.app/repository/database/redis"
)

func CreateUserUseCase(ctx *gin.Context, userEmail string) (*string, *string, error) {
	var userRepoInstance = repository.GetUserRepo()
	cached_user := redis.RedisRepo.FindOne(ctx, fmt.Sprintf("%s-user", userEmail))
	if cached_user == nil {
		err := errors.New("sign up required")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
		return nil, nil, err
	}
	var payload models.User
	json.Unmarshal([]byte(*cached_user), &payload)
	response, err := userRepoInstance.CreateOne(payload)
	if err != nil {
		err := errors.New("something went wrong while creating user")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
		return nil, nil, err
	}
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-user", userEmail))
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-otp", userEmail))
	return response, &payload.UserName, err
}
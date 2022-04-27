package user

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	user "mize.app/app/user/models"
	// userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
	redis "mize.app/repository/database/redis"
)

func CacheUserUseCase(ctx *gin.Context, payload *user.User) (bool, error) {
	// exists := userRepo.UserRepository.FindOneByFilter(ctx, map[string]interface{}{"email": payload.Email, "username": payload.UserName})
	var unexprectedError app_errors.RequestError
	// if exists != nil {
	// 	unexprectedError = app_errors.RequestError{StatusCode: http.StatusConflict, Err: errors.New("user with email or username already exists")}
	// 	app_errors.ErrorHandler(ctx, unexprectedError)
	// 	return false, unexprectedError
	// }
	result := redis.RedisRepo.CreateEntry(ctx, fmt.Sprintf("%s-user", payload.Email), payload, 5*time.Minute)
	if !result {
		unexprectedError = app_errors.RequestError{StatusCode: http.StatusConflict, Err: errors.New("something went wrong while creating user")}
		app_errors.ErrorHandler(ctx, unexprectedError)
		return false, unexprectedError
	}
	return result, nil
}

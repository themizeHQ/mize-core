package user

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	user "mize.app/app/user/models"
	userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
	redis "mize.app/repository/database/redis"
)

func CacheUserUseCase(ctx *gin.Context, payload *user.User) error {
	var userRepoInstance = userRepo.GetUserRepo()
	emailExists, err := userRepoInstance.CountDocs(map[string]interface{}{"email": payload.Email})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return err
	}
	if emailExists != 0 {
		err := errors.New("user with email already exists")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
		return err
	}
	usernameExists, err := userRepoInstance.CountDocs(map[string]interface{}{"userName": payload.UserName})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return err
	}
	if usernameExists != 0 {
		err := errors.New("user with username already exists")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
		return err
	}
	result := redis.RedisRepo.CreateEntry(ctx, fmt.Sprintf("%s-user", payload.Email), payload, 20*time.Minute)
	if !result {
		err := errors.New("something went wrong while creating user")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return err
	}
	return nil
}

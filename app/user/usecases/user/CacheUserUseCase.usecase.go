package user

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	user "mize.app/app/user/models"
	userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
	redis "mize.app/repository/database/redis"
)

func CacheUserUseCase(ctx *gin.Context, payload *user.User) (bool, error) {
	payload.Verified = false
	payload.AppsCreated = []primitive.ObjectID{}
	payload.WorkspaceCreated = []primitive.ObjectID{}
	emailExists := userRepo.UserRepository.FindOneByFilter(ctx, map[string]interface{}{"email": payload.Email})
	usernameExists := userRepo.UserRepository.FindOneByFilter(ctx, map[string]interface{}{"userName": payload.UserName})
	var unexprectedError app_errors.RequestError
	if emailExists != nil {
		unexprectedError = app_errors.RequestError{StatusCode: http.StatusConflict, Err: errors.New("user with email already exists")}
		app_errors.ErrorHandler(ctx, unexprectedError, http.StatusInternalServerError)
		return false, unexprectedError
	}
	if usernameExists != nil {
		unexprectedError = app_errors.RequestError{StatusCode: http.StatusConflict, Err: errors.New("user with username already exists")}
		app_errors.ErrorHandler(ctx, unexprectedError, http.StatusInternalServerError)
		return false, unexprectedError
	}
	result := redis.RedisRepo.CreateEntry(ctx, fmt.Sprintf("%s-user", payload.Email), payload, 20*time.Minute)
	if !result {
		unexprectedError = app_errors.RequestError{StatusCode: http.StatusConflict, Err: errors.New("something went wrong while creating user")}
		app_errors.ErrorHandler(ctx, unexprectedError, http.StatusInternalServerError)
		return false, unexprectedError
	}
	return result, nil
}

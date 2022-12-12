package usecases

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"mize.app/app/auth/types"
	"mize.app/app/user/models"
	"mize.app/app/user/repository"
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/calls/azure"
	"mize.app/emitter"
	"mize.app/repository/database/redis"
)

func CreateUserUseCase(ctx *gin.Context, payload types.VerifyData) (accessToken *string, refreshToken *string, user *models.User) {
	cached_user := redis.RedisRepo.FindOne(ctx, fmt.Sprintf("%s-user", payload.Email))
	if cached_user == nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("this user does not exist"), StatusCode: http.StatusNotFound})
		return nil, nil, nil
	}
	var data models.User
	json.Unmarshal([]byte(*cached_user), &data)
	acsData, err := azure.GenerateUserAndToken()
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return nil, nil, nil
	}
	data.ACSUserId = acsData.User.CommunicationUserId
	var userRepoInstance = repository.GetUserRepo()
	user, err = userRepoInstance.CreateOne(data)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong while creating user"), StatusCode: http.StatusNotFound})
		return nil, nil, nil
	}
	rT, err := authentication.GenerateRefreshToken(ctx, user.Id.Hex(), payload.Email, data.UserName, data.FirstName, data.Language, acsData.User.CommunicationUserId)
	if err != nil {
		return nil, nil, nil
	}
	aT, err := authentication.GenerateAccessToken(ctx, user.Id.Hex(), payload.Email, data.UserName, data.FirstName, data.LastName, nil, acsData.User.CommunicationUserId)
	if err != nil {
		return nil, nil, nil
	}
	emitter.Emitter.Emit(emitter.Events.AUTH_EVENTS.USER_VERIFIED, map[string]string{"email": payload.Email, "firstName": user.FirstName})
	return &aT, &rT, user
}

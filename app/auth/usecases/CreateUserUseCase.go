package usecases

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"mize.app/app/user/models"
	"mize.app/app/user/repository"
	"mize.app/app_errors"
	"mize.app/authentication"
	user_constants "mize.app/constants/user"
	"mize.app/emitter"
)

func CreateUserUseCase(ctx *gin.Context, payload models.User) (accessToken *string, refreshToken *string, user *models.User) {
	var userRepoInstance = repository.GetUserRepo()
	payload.Status = user_constants.AVAILABLE
	if payload.Language == "" {
		payload.Language = "english"
	}
	payload.Discoverability = []user_constants.UserDiscoverability{user_constants.DISCOVERABILITY_EMAIL, user_constants.DISCOVERABILITY_USERNAME, user_constants.DISCOVERABILITY_PHONE}
	user, err := userRepoInstance.CreateOne(payload)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong while creating user"), StatusCode: http.StatusNotFound})
		return nil, nil, nil
	}
	rT, err := authentication.GenerateRefreshToken(ctx, user.Id.Hex(), payload.Email, payload.UserName, payload.FirstName, payload.Language)
	if err != nil {
		return nil, nil, nil
	}
	aT, err := authentication.GenerateAccessToken(ctx, user.Id.Hex(), payload.Email, payload.UserName, payload.FirstName, payload.LastName, nil)
	if err != nil {
		return nil, nil, nil
	}
	emitter.Emitter.Emit(emitter.Events.AUTH_EVENTS.USER_VERIFIED, map[string]string{"email": payload.Email, "firstName": user.FirstName, "id": user.Id.Hex()})
	return &aT, &rT, user
}

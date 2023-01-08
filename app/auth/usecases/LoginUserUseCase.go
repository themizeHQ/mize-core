package usecases

import (
	"errors"
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
	"mize.app/app/auth/types"
	"mize.app/app/user/models"
	"mize.app/app/user/repository"
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/cryptography"
)

func LoginUserUseCase(ctx *gin.Context, payload types.LoginDetails) (refreshToken *string, accessToken *string, user *models.User) {
	var userRepoInstance = repository.GetUserRepo()
	var profile *models.User
	if _, err := mail.ParseAddress(payload.Account); err == nil {
		p, err := userRepoInstance.FindOneByFilter(map[string]interface{}{"email": payload.Account})
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
			return
		}
		profile = p
	} else {
		p, err := userRepoInstance.FindOneByFilter(map[string]interface{}{"userName": payload.Account})
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
			return
		}
		profile = p
	}
	if profile == nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("user does not exist"), StatusCode: http.StatusNotFound})
		return
	}
	if profile.Password == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("sign in using oauth"), StatusCode: http.StatusBadRequest})
		return
	}
	success := cryptography.VerifyData(profile.Password, payload.Password)
	if !success {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("incorrect password"), StatusCode: http.StatusUnauthorized})
		return
	}
	rT, err := authentication.GenerateRefreshToken(ctx, profile.Id.Hex(), profile.Email, profile.UserName, profile.FirstName, profile.LastName, profile.ACSUserId)
	if err != nil {
		return
	}
	aT, err := authentication.GenerateAccessToken(ctx, profile.Id.Hex(), profile.Email, profile.UserName, profile.FirstName, profile.LastName, nil, profile.ACSUserId)
	if err != nil {
		return
	}
	profile.Password = ""
	return &rT, &aT, profile
}

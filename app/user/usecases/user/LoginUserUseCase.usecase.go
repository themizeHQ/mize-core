package user

import (
	"errors"
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"

	user "mize.app/app/user/models"
	userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
	"mize.app/cryptography"
)

func LoginUserUseCase(ctx *gin.Context, account string, password string) *user.User {
	var userRepoInstance = userRepo.GetUserRepo()
	var profile *user.User
	if _, err := mail.ParseAddress(account); err == nil {
		p, err := userRepo.UserRepository.FindOneByFilter(map[string]interface{}{"email": account})
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return nil
		}
		profile = p
	} else {
		p, err := userRepoInstance.FindOneByFilter(map[string]interface{}{"userName": account})
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return nil
		}
		profile = p
	}
	if profile == nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("user does not exist"), StatusCode: http.StatusNotFound})
		return nil
	}
	success := cryptography.VerifyData(profile.Password, password)
	if !success {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("incorrect password"), StatusCode: http.StatusUnauthorized})
		return nil
	}
	return profile
}

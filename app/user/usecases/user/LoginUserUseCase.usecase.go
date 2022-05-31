package user

import (
	"net/mail"

	"github.com/gin-gonic/gin"

	user "mize.app/app/user/models"
	userRepo "mize.app/app/user/repository"
	"mize.app/cryptography"
)

func LoginUserUseCase(ctx *gin.Context, account string, password string) (bool, *user.User) {
	profile := getProfile(ctx, account)
	if profile == nil {
		return false, nil
	}
	success := cryptography.VerifyData(profile.Password, password)
	return success, profile
}

func getProfile(ctx *gin.Context, account string) *user.User {
	var profile *user.User
	if _, err := mail.ParseAddress(account); err == nil {
		profile = userRepo.UserRepository.FindOneByFilter(ctx, map[string]interface{}{
			"email": account,
		})
	} else {
		profile = userRepo.UserRepository.FindOneByFilter(ctx, map[string]interface{}{
			"userName": account,
		})
	}
	return profile
}

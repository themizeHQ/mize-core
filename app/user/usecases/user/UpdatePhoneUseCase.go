package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
)

func UpdatePhoneUseCase(ctx *gin.Context, phone string) bool {
	userRepoInstance := userRepo.GetUserRepo()
	success, err := userRepoInstance.UpdatePartialById(ctx, ctx.GetString("UserId"), map[string]string{"phone": phone})
	if err != nil || !success {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return false
	}
	return success
}

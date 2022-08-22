package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/user/models"
	userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
)

func UpdateUserUseCase(ctx *gin.Context, payload models.UpdateUser) bool {
	userRepoInstance := userRepo.GetUserRepo()
	success, err := userRepoInstance.UpdatePartialById(ctx, ctx.GetString("UserId"), payload)
	if err != nil || !success {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return false
	}
	return success
}

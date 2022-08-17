package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
)

func UpdateUserUseCase(ctx *gin.Context, payload map[string]interface{}) bool {
	userRepoInstance := userRepo.GetUserRepo()
	success, err := userRepoInstance.UpdatePartialById(ctx, ctx.GetString("UserId"),
		deleteKeys(payload, []string{"password", "_id", "verified", "email", "createdAt", "updatedAt"}))
	if err != nil || !success {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return false
	}
	return success
}

func deleteKeys(payload map[string]interface{}, keys []string) map[string]interface{} {
	for _, k := range keys {
		delete(payload, k)
	}
	return payload
}

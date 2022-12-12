package usecases

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mize.app/app/auth/types"
	"mize.app/app/user/repository"
	"mize.app/app_errors"
	"mize.app/cryptography"
)

func UpdatePasswordUseCase(ctx *gin.Context, payload types.UpdatePassword) error {
	userRepoInstance := repository.GetUserRepo()
	user, err := userRepoInstance.FindById(ctx.GetString("UserId"), options.FindOne().SetProjection(map[string]int{
		"password": 1,
	}))
	if err != nil {
		err = errors.New("user does not exist")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
		return err
	}
	success := cryptography.VerifyData(user.Password, payload.CurrentPassword)
	if !success {
		err = errors.New("incorrect password")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return err
	}
	user.Password = payload.NewPassword
	err = user.ValidatePassword()
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	success, err = userRepoInstance.UpdatePartialById(ctx, ctx.GetString("UserId"), map[string]interface{}{
		"password": string(cryptography.HashString(user.Password, nil)),
	})
	if !success || err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong while updating password"), StatusCode: http.StatusInternalServerError})
	}
	return nil
}

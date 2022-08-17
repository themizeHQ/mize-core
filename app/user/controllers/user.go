package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	userRepo "mize.app/app/user/repository"
	userUseCases "mize.app/app/user/usecases/user"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func FetchProfile(ctx *gin.Context) {
	userRepo := userRepo.GetUserRepo()
	profile, err := userRepo.FindById(ctx.GetString("UserId"), options.FindOne().SetProjection(map[string]int{
		"password": 0,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return
	}
	server_response.Response(ctx, http.StatusOK, "profile fetched", true, profile)
}

func FetchUsersProfile(ctx *gin.Context) {
	userRepo := userRepo.GetUserRepo()
	id := ctx.Params.ByName("id")
	if id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the user id"), StatusCode: http.StatusBadRequest})
		return
	}
	profile, err := userRepo.FindById(id, options.FindOne().SetProjection(map[string]int{
		"password": 0,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return
	}
	profile.Password = ""
	server_response.Response(ctx, http.StatusOK, "profile fetched", true, profile)
}

func UpdateUserData(ctx *gin.Context) {
	var payload map[string]interface{}
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json with valid fields"), StatusCode: http.StatusBadRequest})
		return
	}
	success := userUseCases.UpdateUserUseCase(ctx, payload)
	if !success {
		return
	}
	server_response.Response(ctx, http.StatusOK, "profile updated", success, nil)
}

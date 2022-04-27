package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	userModel "mize.app/app/user/models"
	useCases "mize.app/app/user/usecases/user"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func CacheUser(ctx *gin.Context) {
	var payload userModel.User
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, err)
		return
	}
	response, err := useCases.CacheUserUseCase(ctx, &payload)
	if err != nil {
		server_response.Response(ctx, http.StatusInternalServerError, "Something went wrong while creating user", false, nil)
		return
	}
	server_response.Response(ctx, http.StatusCreated, "User created successfuly", true, response)
}

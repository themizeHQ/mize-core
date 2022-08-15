package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/server_response"
	"mize.app/utils"
)

func FetchUserWorkspaces(ctx *gin.Context) {
	workspaceMemberRepo := repository.GetWorkspaceMember()
	payload, err := workspaceMemberRepo.FindMany(map[string]interface{}{"userId": utils.HexToMongoId(ctx, ctx.GetString("UserId"))})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: app_errors.RequestError{StatusCode: http.StatusInternalServerError,
			Err: errors.New("could not retrieve your workspaces at this time")}, StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "workspaces retrieved", true, payload)
}

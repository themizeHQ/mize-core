package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/server_response"
	"mize.app/utils"
)

func FetchUserWorkspaces(ctx *gin.Context) {
	workspaceMemberRepo := repository.GetWorkspaceMember()
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil || page == 0 {
		page = 1
	}
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 15
	}
	skip := (page - 1) * limit
	payload, err := workspaceMemberRepo.FindMany(map[string]interface{}{
		"userId": utils.HexToMongoId(ctx, ctx.GetString("UserId")),
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	},
	)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: app_errors.RequestError{StatusCode: http.StatusInternalServerError,
			Err: errors.New("could not retrieve your workspaces at this time")}, StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "workspaces retrieved", true, payload)
}

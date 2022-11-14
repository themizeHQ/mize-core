package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	payload, err := workspaceMemberRepo.FindManyStripped(map[string]interface{}{
		"userId": utils.HexToMongoId(ctx, ctx.GetString("UserId")),
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}, options.Find().SetProjection(map[string]int{
		"workspaceName": 1,
		"profileImage":  1,
		"workspaceId":   1,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: app_errors.RequestError{StatusCode: http.StatusInternalServerError,
			Err: errors.New("could not retrieve your workspaces at this time")}, StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "workspaces retrieved", true, payload)
}

func FetchWorkspacesMembers(ctx *gin.Context) {
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
	payload, err := workspaceMemberRepo.FindManyStripped(map[string]interface{}{
		"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}, options.Find().SetProjection(map[string]int{
		"profileImage":  1,
		"firstName":  1,
		"lastName":  1,
		"admin":  1,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: app_errors.RequestError{StatusCode: http.StatusInternalServerError,
			Err: errors.New("could not retrieve your workspaces at this time")}, StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "workspaces retrieved", true, payload)
}

func SearchWorkspaceMembers(ctx *gin.Context) {
	term := ctx.Query("term")
	if strings.TrimSpace(term) == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a search term"), StatusCode: http.StatusBadRequest})
		return
	}
	workspaceRepo := repository.GetWorkspaceMember()
	profile, err := workspaceRepo.FindManyStripped(map[string]interface{}{
		"$or": []map[string]interface{}{
			{
				"userName":    map[string]interface{}{"$regex": term, "$options": "im"},
				"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			},
			{
				"firstName":   map[string]interface{}{"$regex": term, "$options": "im"},
				"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			},
			{
				"lastName":    map[string]interface{}{"$regex": term, "$options": "im"},
				"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			},
		},
	}, options.Find().SetProjection(map[string]int{
		"firstName":    1,
		"lastName":    1,
		"userName":     1,
		"profileImage": 1,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not complete search"), StatusCode: http.StatusInternalServerError})
		return
	}
	server_response.Response(ctx, http.StatusOK, "search complete", true, profile)
}

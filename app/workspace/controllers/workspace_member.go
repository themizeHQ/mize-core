package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/workspace"
	"mize.app/app/workspace/repository"
	"mize.app/app/workspace/usecases/workspace_member"
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/server_response"
	"mize.app/utils"
)

func FetchUserWorkspaces(ctx *gin.Context) {
	workspaceMemberRepo := repository.GetWorkspaceMember()
	admin := ctx.Query("admin")
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
		"admin": func() interface{} {
			if admin != "" {
				adminBool, err := strconv.ParseBool(admin)
				if err != nil {
					return map[string]interface{}{
						"$in": []bool{true, false},
					}
				}
				return adminBool
			}
			return map[string]interface{}{
				"$in": []bool{true, false},
			}
		}(),
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
	admin := ctx.Query("admin")
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
	}, options.Find().SetProjection(func() map[string]int {
		if admin == "true" && authentication.RoleType(ctx.GetString("Role")) == authentication.ADMIN {
			return map[string]int{
				"profileImage": 1,
				"firstName":    1,
				"lastName":     1,
				"userName":     1,
				"email":        1,
				"phone":        1,
				"admin":        1,
			}
		}
		return map[string]int{
			"profileImage": 1,
			"firstName":    1,
			"lastName":     1,
			"admin":        1,
		}
	}()))
	if err != nil {
		fmt.Println(err)
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: app_errors.RequestError{StatusCode: http.StatusInternalServerError,
			Err: errors.New("could not retrieve your workspaces at this time")}, StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "workspaces retrieved", true, payload)
}

func SearchWorkspaceMembers(ctx *gin.Context) {
	term := ctx.Query("term")
	admin := ctx.Query("admin")
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil || page == 0 {
		page = 1
	}
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 15
	}
	skip := (page - 1) * limit
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
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}, options.Find().SetProjection(func() map[string]int {
		if admin == "true" && authentication.RoleType(ctx.GetString("Role")) == authentication.ADMIN {
			return map[string]int{
				"profileImage": 1,
				"firstName":    1,
				"lastName":     1,
				"userName":     1,
				"email":        1,
				"phone":        1,
				"admin":        1,
			}
		}
		return map[string]int{
			"profileImage": 1,
			"firstName":    1,
			"lastName":     1,
			"userName":     1,
		}
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not complete search"), StatusCode: http.StatusInternalServerError})
		return
	}
	server_response.Response(ctx, http.StatusOK, "search complete", true, profile)
}

func AddAdminAccess(ctx *gin.Context) {
	var payload workspace.AddWorkspacePiviledges
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the appropraite payload"), StatusCode: http.StatusBadRequest})
		return
	}
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	success := workspace_member.AddAdminAccessUseCase(ctx, payload.Id, payload.Permissions)
	if !success {
		return
	}
	server_response.Response(ctx, http.StatusOK, "priviledges granted", true, nil)
}

func DeactivateWorkspaceMember(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a workspace member id"), StatusCode: http.StatusBadRequest})
		return
	}
	success := workspace_member.DeactivateWorkspaceMemberUseCase(ctx, id)
	if !success {
		return
	}
	server_response.Response(ctx, http.StatusOK, "member deactivated", true, nil)
}

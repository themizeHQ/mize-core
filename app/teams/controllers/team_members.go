package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mize.app/app/teams/repository"
	"mize.app/app/teams/types"
	teammembersUseCases "mize.app/app/teams/usecases/team_members"
	"mize.app/app_errors"
	"mize.app/server_response"
	"mize.app/utils"
)

// CreateTeamMembers - controller function to create multiple team members
func CreateTeamMembers(ctx *gin.Context) {
	var payload []types.TeamMembersPayload
	var teamID = ctx.Params.ByName("id")
	if err := ctx.ShouldBind(&payload); err != nil {
		server_response.Response(ctx, http.StatusBadRequest, "pass in correct payload", false, nil)
		return
	}
	err := teammembersUseCases.CreateTeamMemberUseCase(ctx, teamID, payload)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "team members created", true, nil)
}

func FetchTeamMembers(ctx *gin.Context) {
	teamID := ctx.Params.ByName("id")
	if teamID == "" {
		server_response.Response(ctx, http.StatusBadRequest, "pass in teamID", false, nil)
		return
	}
	teamMembersRepo := repository.GetTeamMemberRepo()
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil || page == 0 {
		page = 1
	}
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 15
	}
	skip := (page - 1) * limit
	members, err := teamMembersRepo.FindManyStripped(map[string]interface{}{
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		"teamId":      *utils.HexToMongoId(ctx, teamID),
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}, options.Find().SetProjection(
		map[string]interface{}{
			"firstName":    1,
			"lastName":     1,
			"profileImage": 1,
		},
	))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: app_errors.RequestError{StatusCode: http.StatusInternalServerError,
			Err: errors.New("could not retrieve your channels at this time")}, StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusOK, "team members retrieved", true, members)
}

func RemoveTeamMembers(ctx *gin.Context) {
	var payload types.IDArray
	if err := ctx.ShouldBind(&payload); err != nil {
		server_response.Response(ctx, http.StatusBadRequest, "pass in an array of team members to remove", false, nil)
		return
	}
	teammembersUseCases.RemoveTeamMembers(ctx, payload)
	server_response.Response(ctx, http.StatusCreated, "team members removed", true, nil)
}

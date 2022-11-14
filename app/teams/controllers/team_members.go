package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mize.app/app/teams/types"
	teammembersUseCases "mize.app/app/teams/usecases/team_members"
	"mize.app/server_response"
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

func RemoveTeamMembers(ctx *gin.Context) {
	var payload types.IDArray
	if err := ctx.ShouldBind(&payload); err != nil {
		server_response.Response(ctx, http.StatusBadRequest, "pass in an array of team members to remove", false, nil)
		return
	}
	teammembersUseCases.RemoveTeamMembers(ctx, payload)
	server_response.Response(ctx, http.StatusCreated, "team members removed", true, nil)
}

package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/workspace/repository"
	workspaceUseCases "mize.app/app/workspace/usecases/workspace"
	workspaceInviteUseCases "mize.app/app/workspace/usecases/workspace_invite"
	workspaceMemberUseCases "mize.app/app/workspace/usecases/workspace_member"
	"mize.app/app_errors"
	"mize.app/server_response"
	"mize.app/utils"
)

func InviteToWorkspace(ctx *gin.Context) {
	var payload struct {
		Emails []string
	}
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json value"), StatusCode: http.StatusBadRequest})
		return
	}
	if len(payload.Emails) == 0 {
		server_response.Response(ctx, http.StatusBadRequest, "Pass in an array of emails to get invites", false, nil)
		return
	}
	err := workspaceUseCases.SendInvitesUseCase(ctx, payload.Emails)
	if err != nil {
		return
	}

	server_response.Response(ctx, http.StatusOK, fmt.Sprintf("invites sent successfully to %d emails", len(payload.Emails)), true, nil)
}

func RejectWorkspaceInvite(ctx *gin.Context) {
	inviteId := ctx.Params.ByName("inviteId")
	if inviteId == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("workspace invite id was not provided"),
			StatusCode: http.StatusBadRequest})
		return
	}
	success, err := workspaceInviteUseCases.RejectWorkspaceInviteUseCase(ctx, inviteId)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "invite rejected", true, success)
}

func AcceptWorkspaceInvite(ctx *gin.Context) {
	inviteId := ctx.Params.ByName("inviteId")
	if inviteId == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("workspace invite id was not provided"),
			StatusCode: http.StatusBadRequest})
		return
	}
	workspace_id, err := workspaceInviteUseCases.AcceptWorkspaceInviteUseCase(ctx, inviteId)
	if err != nil {
		return
	}
	id, err := workspaceMemberUseCases.CreateWorkspaceMemberUseCase(ctx, *workspace_id, false)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("failed to join workspace"),
			StatusCode: http.StatusInternalServerError})
		return
	}
	server_response.Response(ctx, http.StatusOK, "invite accepted. you have been added to the workspace", true, id)
}

func FetchWorkspaceInvites(ctx *gin.Context) {
	workspaceInvitesRepo := repository.GetWorkspaceInviteRepo()
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil || page == 0 {
		page = 1
	}
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 15
	}
	skip := (page - 1) * limit
	invites, err := workspaceInvitesRepo.FindMany(map[string]interface{}{
		"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	})
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "invites fetched", true, invites)
}

package workspace

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	workspaceUseCases "mize.app/app/workspace/usecases/workspace"
	workspaceInviteUseCases "mize.app/app/workspace/usecases/workspace_invite"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func InviteToWorkspace(ctx *gin.Context) {
	var payload struct {
		Emails      []string
		WorkspaceId string
	}
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json value"), StatusCode: http.StatusBadRequest})
		return
	}
	if len(payload.Emails) == 0 {
		server_response.Response(ctx, http.StatusBadRequest, "Pass in an array of emails to get invites", false, nil)
		return
	}
	if payload.WorkspaceId == "" {
		server_response.Response(ctx, http.StatusBadRequest, "Pass in the workspace name", false, nil)
		return
	}
	err := workspaceUseCases.SendInvitesUseCase(ctx, payload.Emails, payload.WorkspaceId)
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
	success, err := workspaceInviteUseCases.AcceptWorkspaceInviteUseCase(ctx, inviteId)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "invite accepted", true, success)
}

package workspace

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	workspaceUseCases "mize.app/app/workspace/usecases/workspace"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func InviteToWorkspace(ctx *gin.Context) {
	var payload struct {
		Emails        []string
		WorkspaceName string
	}
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return
	}
	if len(payload.Emails) == 0 {
		server_response.Response(ctx, http.StatusBadRequest, "Pass in an array of emails to get invites", false, nil)
		return
	}
	if payload.WorkspaceName == "" {
		server_response.Response(ctx, http.StatusBadRequest, "Pass in the workspace name", false, nil)
		return
	}
	err := workspaceUseCases.SendInvitesUseCase(ctx, payload.Emails, payload.WorkspaceName)
	if err != nil {
		return
	}

	server_response.Response(ctx, http.StatusOK, fmt.Sprintf("Invites sent successfully to %d emails", len(payload.Emails)), true, err)
}

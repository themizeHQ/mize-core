package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/workspace/models"
	workspaceUseCases "mize.app/app/workspace/usecases/workspace"
	workspaceMemberUseCases "mize.app/app/workspace/usecases/workspace_member"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func CreateWorkspace(ctx *gin.Context) {
	var payload models.Workspace
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json value"), StatusCode: http.StatusBadRequest})
		return
	}
	id, err := workspaceUseCases.CreateWorkspaceUseCase(ctx, payload)
	if err != nil {
		return
	}
	member_id, err := workspaceMemberUseCases.CreateWorkspaceMemberUseCase(ctx, *id, true)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "workspace successfully created.", true, map[string]interface{}{
		"workspace_id":        id,
		"workspace_member_id": member_id,
	})
}

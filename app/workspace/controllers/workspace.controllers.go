package workspace

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	workspaceModel "mize.app/app/workspace/models"
	workspaceUseCases "mize.app/app/workspace/usecases/workspace"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func CreateWorkspace(ctx *gin.Context) {
	var payload workspaceModel.Workspace
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json value"), StatusCode: http.StatusBadRequest})
		return
	}
	id, err := workspaceUseCases.CreateWorkspaceUseCase(ctx, payload)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "workspace successfully created.", true, id)
}

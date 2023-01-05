package workspace_member

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	workspace_member_constants "mize.app/constants/workspace"
	"mize.app/logger"
)

func AddAdminAccessUseCase(ctx *gin.Context, id string, permissions []workspace_member_constants.AdminAccessType) bool {
	workspaceMemberRepo := repository.GetWorkspaceMember()
	workspaceRepo := repository.GetWorkspaceRepo()
	member, err := workspaceMemberRepo.FindOneByFilter(map[string]interface{}{
		"_id": id,
	}, options.FindOne().SetProjection(map[string]interface{}{
		"userId": 1,
	}))
	if err != nil {
		e := errors.New("error checking if a user exists")
		logger.Error(e, zap.Error(err))
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: e, StatusCode: http.StatusInternalServerError})
		return false
	}
	if member == nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("user is not a member of this workspace"), StatusCode: http.StatusNotFound})
		return false
	}

	workspace, err := workspaceRepo.FindOneByFilter(map[string]interface{}{
		"_id": ctx.GetString("Workspace"),
	}, options.FindOne().SetProjection(map[string]interface{}{
		"createdBy": 1,
	}))
	if err != nil {
		e := errors.New("error fetching workspace")
		logger.Error(e, zap.Error(err))
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: e, StatusCode: http.StatusInternalServerError})
		return false
	}
	if workspace == nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("workspace does not exist"), StatusCode: http.StatusNotFound})
		return false
	}
	if workspace.CreatedBy.Hex() == member.UserId.Hex() {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you cannot modify permissions for this user"), StatusCode: http.StatusUnauthorized})
		return false
	}
	success, err := workspaceMemberRepo.UpdatePartialById(ctx, id, map[string]interface{}{
		"admin":       len(permissions) != 0,
		"adminAccess": permissions,
	})
	if err != nil || !success {
		e := errors.New("could not update workspace admin priviledges")
		logger.Error(e, zap.Error(err))
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: e, StatusCode: http.StatusNotFound})
		return false
	}
	return success
}

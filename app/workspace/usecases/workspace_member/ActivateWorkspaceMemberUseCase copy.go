package workspace_member

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/logger"
	"mize.app/utils"
)

func ActivateWorkspaceMemberUseCase(ctx *gin.Context, id string) bool {
	workspaceMemberRepo := repository.GetWorkspaceMember()
	workspaceRepo := repository.GetWorkspaceRepo()
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
	member, err := workspaceMemberRepo.FindOneByFilter(map[string]interface{}{
		"_id": id,
	}, options.FindOne().SetProjection(map[string]interface{}{
		"userId": 1,
	}))
	if err != nil {
		e := errors.New("error fetching workspace member")
		logger.Error(e, zap.Error(err))
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: e, StatusCode: http.StatusInternalServerError})
		return false
	}
	if member == nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("user is not a member of this workspace"), StatusCode: http.StatusNotFound})
		return false
	}
	if workspace.CreatedBy.Hex() == member.UserId.Hex() {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you cannot modify permissions for this user"), StatusCode: http.StatusUnauthorized})
		return false
	}
	success, err := workspaceMemberRepo.UpdatePartialByFilter(ctx, map[string]interface{}{
		"_id":         *utils.HexToMongoId(ctx, id),
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	}, map[string]interface{}{
		"deactivated": false,
	})
	if err != nil || !success {
		e := errors.New("could not activate workspace member")
		logger.Error(e, zap.Error(err))
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: e, StatusCode: http.StatusNotFound})
		return false
	}
	return success
}

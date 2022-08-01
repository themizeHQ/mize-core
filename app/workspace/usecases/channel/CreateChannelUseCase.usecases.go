package workspace

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func CreateChannelUseCase(ctx *gin.Context, payload []workspaceModel.Channel) (*[]string, error) {
	var channelRepoInstance = workspaceRepo.GetChannelRepo()
	var workspaceMemberRepoInstance = workspaceRepo.GetWorkspaceMember()
	workspace_member, err := workspaceMemberRepoInstance.FindOneByFilter(map[string]interface{}{
		"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		"userId":      utils.HexToMongoId(ctx, ctx.GetString("UserId")),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("workspace does not exist"), StatusCode: http.StatusNotFound})
		return nil, err
	}
	if !workspace_member.Admin {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("only admins can create channels"), StatusCode: http.StatusUnauthorized})
		return nil, err
	}
	populated_payload := []workspaceModel.Channel{}
	for _, ch_name := range payload {
		ch_name.CreatedBy = *utils.HexToMongoId(ctx, ctx.GetString("UserId"))
		ch_name.WorkspaceId = *utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
		populated_payload = append(populated_payload, ch_name)
	}
	ids, err := channelRepoInstance.CreateBulk(populated_payload)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("channel creation failed"), StatusCode: http.StatusInternalServerError})
		return nil, err
	}
	return ids, nil
}

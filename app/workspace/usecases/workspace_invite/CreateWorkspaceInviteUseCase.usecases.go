package workspace

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	notificationModels "mize.app/app/notification/models"
	notificationUseCases "mize.app/app/notification/usecases"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	constnotif "mize.app/constants/notification"
	"mize.app/utils"
)

func CreateWorkspaceInviteUseCase(ctx *gin.Context, workspace_name string, filter map[string]interface{}, payload map[string]interface{}) error {
	var workspaceInviteRepoInstance = workspaceRepo.GetWorkspaceInviteRepo()
	inviteId, err := workspaceInviteRepoInstance.UpdateOrCreateByFieldAndReturn(filter, payload, options.Update().SetUpsert(true))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	if inviteId == nil {
		return nil
	}
	notificationUseCases.CreateNotificationUseCase(ctx, notificationModels.Notification{
		WorkspaceId: *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		ResourceId:  *utils.HexToMongoId(ctx, *inviteId),
		Importance:  constnotif.NOTIFICATION_NORMAL,
		Type:        constnotif.WORKSPACE_INVITE,
		Scope:       constnotif.USER_NOTIFICATION,
		Message:     fmt.Sprintf("You have been invited to join the %s workspace", workspace_name),
	})
	return nil
}

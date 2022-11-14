package workspace_invite

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	notificationModels "mize.app/app/notification/models"
	notificationUseCases "mize.app/app/notification/usecases"
	userRepo "mize.app/app/user/repository"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	notification_constants "mize.app/constants/notification"
	"mize.app/utils"
)

func CreateWorkspaceInviteUseCase(ctx *gin.Context, filter map[string]interface{}, payload map[string]interface{}) error {
	var workspaceInviteRepoInstance = workspaceRepo.GetWorkspaceInviteRepo()
	inviteId, err := workspaceInviteRepoInstance.UpdateOrCreateByFieldAndReturn(filter, payload, options.Update().SetUpsert(true))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	if inviteId == nil {
		return nil
	}
	userRepo := userRepo.GetUserRepo()
	user, err := userRepo.FindOneByFilter(map[string]interface{}{"email": filter["email"]},
		options.FindOne().SetProjection(map[string]int{
			"_id": 1,
		}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	if user == nil {
		return nil
	}
	notificationUseCases.CreateNotificationUseCase(ctx, notificationModels.Notification{
		WorkspaceId: nil,
		UserId:      utils.HexToMongoId(ctx, user.Id.Hex()),
		ResourceId:  *utils.HexToMongoId(ctx, *inviteId),
		Importance:  notification_constants.NOTIFICATION_NORMAL,
		Type:        notification_constants.WORKSPACE_INVITE,
		Scope:       notification_constants.USER_NOTIFICATION,
		Message:     fmt.Sprintf("You have been invited to join the %s workspace", ctx.GetString("WorkspaceName")),
	})
	return nil
}

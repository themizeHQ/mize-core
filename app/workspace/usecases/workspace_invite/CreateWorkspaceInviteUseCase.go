package workspace_invite

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	notificationModels "mize.app/app/notification/models"
	notificationUseCases "mize.app/app/notification/usecases"
	userRepo "mize.app/app/user/repository"
	workspaceModels "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	notification_constants "mize.app/constants/notification"
	"mize.app/utils"
)

func CreateWorkspaceInviteUseCase(ctx *gin.Context, filter map[string]interface{}, payload workspaceModels.WorkspaceInvite) error {
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
	reacted := false
	notificationUseCases.CreateNotificationUseCase(ctx, notificationModels.Notification{
		UserId:     utils.HexToMongoId(ctx, user.Id.Hex()),
		ResourceId: utils.HexToMongoId(ctx, *inviteId),
		Importance: notification_constants.NOTIFICATION_NORMAL,
		ImageURL:   filter["image"].(string),
		Type:       notification_constants.WORKSPACE_INVITE,
		Scope:      notification_constants.USER_NOTIFICATION,
		Header:     fmt.Sprintf("%s added you to %s", ctx.GetString("Firstname"), ctx.GetString("WorkspaceName")),
		Message:    fmt.Sprintf("Accept or Reject your invitation to %s", ctx.GetString("WorkspaceName")),
		Reacted:    &reacted,
	})
	return nil
}

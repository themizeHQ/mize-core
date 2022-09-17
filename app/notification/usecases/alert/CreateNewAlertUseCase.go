package alert

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/notification/models"
	"mize.app/app/notification/repository"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	notification_constants "mize.app/constants/notification"
	workspace_constants "mize.app/constants/workspace"
	"mize.app/realtime"
	"mize.app/utils"
)

type UserPayload struct {
	Message    string
	Importance string
	ResourceId *string
	UserIds    []string
}

func CreateNewAlertUseCase(ctx *gin.Context, payload UserPayload) bool {
	workspaceMemberRepo := workspaceRepo.GetWorkspaceMember()
	admin, err := workspaceMemberRepo.FindOneByFilter(map[string]interface{}{
		"userId":      utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	}, options.FindOne().SetProjection(
		map[string]int{
			"adminAccess": 1,
		},
	))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("unable to send alert"), StatusCode: http.StatusInternalServerError})
		return false
	}
	has_access := admin.HasAccess([]workspace_constants.AdminAccessType{
		workspace_constants.AdminAccess.FULL_ACCESS,
	})
	if !has_access {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you do not have this permission"), StatusCode: http.StatusUnauthorized})
		return false
	}
	alertRepo := repository.GetAlertRepo()
	_, err = alertRepo.CreateOne(models.Alert{
		Message:     payload.Message,
		Importance:  notification_constants.NotificationImportanceLevel(payload.Importance),
		ResourceId:  parseResourceId(ctx, payload.ResourceId),
		UserIds:     *parseUserIds(ctx, payload.UserIds),
		AdminId:     *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		WorkspaceId: *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("unable to send alert"), StatusCode: http.StatusInternalServerError})
		return false
	}
	for _, id := range payload.UserIds {
		realtime.CentrifugoController.Publish(id, map[string]interface{}{
			"time":       time.Now(),
			"resourceId": payload.ResourceId,
			"importance": payload.Importance,
			"message":    payload.Message,
		})
	}
	return true
}

func parseResourceId(ctx *gin.Context, id *string) *primitive.ObjectID {
	if id == nil {
		return nil
	} else {
		return utils.HexToMongoId(ctx, *id)
	}
}

func parseUserIds(ctx *gin.Context, ids []string) *[]primitive.ObjectID {
	parsedIds := []primitive.ObjectID{}
	for _, id := range ids {
		parsedIds = append(parsedIds, *utils.HexToMongoId(ctx, id))
	}
	return &parsedIds
}

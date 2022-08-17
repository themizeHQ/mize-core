package alert

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"mize.app/app/notification/models"
	"mize.app/app/notification/repository"
	"mize.app/app_errors"
	notification_constants "mize.app/constants/notification"
	"mize.app/utils"
)

type UserPayload struct {
	Message    string
	Importance string
	ResourceId *string
	UserIds    []string
}

func CreateNewAlertUseCase(ctx *gin.Context, payload UserPayload) bool {
	alertRepo := repository.GetAlertRepo()
	_, err := alertRepo.CreateOne(models.Alert{
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

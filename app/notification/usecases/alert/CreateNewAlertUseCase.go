package alert

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/notification/models"
	"mize.app/app/notification/repository"
	userRepository "mize.app/app/user/repository"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	notification_constants "mize.app/constants/notification"
	workspace_constants "mize.app/constants/workspace"
	"mize.app/emails"
	"mize.app/emitter"
	"mize.app/realtime"
	"mize.app/utils"
)

func CreateNewAlertUseCase(ctx *gin.Context, payload models.Alert) bool {
	workspaceMemberRepo := workspaceRepo.GetWorkspaceMember()
	admin, err := workspaceMemberRepo.FindOneByFilter(map[string]interface{}{
		"userId":      utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	}, options.FindOne().SetProjection(
		map[string]int{
			"adminAccess":           1,
			"profileImageThumbnail": 1,
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
	payload.AdminName = fmt.Sprintf("%s %s", ctx.GetString("Firstname"), ctx.GetString("Lastname"))
	payload.ImageURL = *admin.ProfileImageThumbNail
	alertRepo := repository.GetAlertRepo()
	_, err = alertRepo.CreateOne(models.Alert{
		Message:      payload.Message,
		Importance:   notification_constants.NotificationImportanceLevel(payload.Importance),
		ResourceId:   payload.ResourceId,
		UserIds:      payload.UserIds,
		AdminId:      *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		WorkspaceId:  *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		AlertByEmail: payload.AlertByEmail,
		AlertBySMS:   payload.AlertBySMS,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("unable to send alert"), StatusCode: http.StatusInternalServerError})
		return false
	}
	var wg sync.WaitGroup
	sendInAppNotifications(payload)
	userRepo := userRepository.GetUserRepo()
	if payload.AlertByEmail {
		for _, id := range payload.UserIds {
			wg.Add(1)
			go func(id string) {
				defer func() {
					wg.Done()
				}()
				user, err := userRepo.FindById(id)
				if err != nil {
					return
				}
				if user == nil {
					return
				}
				emails.SendEmail(user.Email, fmt.Sprintf("Mize alert from %s", fmt.Sprintf("%s %s", ctx.GetString("Firstname"), ctx.GetString("Lastname"))), "alert_reminder", map[string]interface{}{
					"TIME":    time.Unix(payload.CreatedAt.Time().Unix(), 0).Local().Format(time.UnixDate),
					"DETAILS": payload.Message,
					"NAME":    fmt.Sprintf("You have a new alert from %s.", ctx.GetString("WorkspaceName")),
					"FROM":    fmt.Sprintf("%s %s", ctx.GetString("Firstname"), ctx.GetString("Lastname")),
				})

			}(id.Hex())
		}
		wg.Wait()
	}
	if payload.AlertBySMS && (payload.Importance == notification_constants.NOTIFICATION_IMPORTANT || payload.Importance == notification_constants.NOTIFICATION_VERY_IMPORTANT) {
		for _, id := range payload.UserIds {
			wg.Add(1)
			go func(id string) {
				defer func() {
					wg.Done()
				}()
				user, err := userRepo.FindOneByFilter(map[string]interface{}{
					"_id":   id,
					"phone": map[string]interface{}{"$ne": nil},
				}, options.FindOne().SetProjection(map[string]interface{}{
					"firstName": 1,
					"phone":     1,
				}))
				if err != nil || user == nil {
					return
				}
				emitter.Emitter.Emit(emitter.Events.SMS_EVENTS.SMS_SENT, map[string]interface{}{
					"to":      *user.Phone,
					"message": fmt.Sprintf("Hi %s, you have a new %s alert from %s on the %s workspace.", user.FirstName, payload.Importance, fmt.Sprintf("%s %s", ctx.GetString("Firstname"), ctx.GetString("Lastname")), ctx.GetString("WorkspaceName")),
				})
			}(id.Hex())
		}
		wg.Wait()
	}
	return true
}

func sendInAppNotifications(payload models.Alert) {
	var wg sync.WaitGroup
	for _, id := range payload.UserIds {
		wg.Add(1)
		go func(i string) {
			defer func() {
				wg.Done()
			}()
			realtime.CentrifugoController.Publish(i, realtime.MessageScope.ALERT, map[string]interface{}{
				"createdAt":  payload.CreatedAt,
				"resourceId": payload.ResourceId,
				"importance": payload.Importance,
				"message":    payload.Message,
			})
		}(id.Hex())
	}
	wg.Wait()
}

package notification

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	notificationModel "mize.app/app/notification/models"
	notification "mize.app/app/notification/repository"
	notificationUseCases "mize.app/app/notification/usecases"
	"mize.app/app_errors"
	constnotif "mize.app/constants/notification"
	"mize.app/server_response"
	"mize.app/utils"
)

func FetchUserNotifications(ctx *gin.Context) {
	notificationRepo := notification.GetNotificationRepo()
	var notifications *[]notificationModel.Notification
	var err error
	if ctx.GetString("Workspace") != "" {
		notifications, err = notificationRepo.FindMany(map[string]interface{}{
			"userId": map[string]interface{}{
				"$in": []interface{}{utils.HexToMongoId(ctx, ctx.GetString("UserId")), nil},
			},
			"scope": map[string]interface{}{
				"$in": []constnotif.NotificationScope{
					constnotif.APP_WIDE_NOTIFICATION,
					constnotif.USER_NOTIFICATION,
				},
			},
		})
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return
		}
	} else {
		notifications, err = notificationRepo.FindMany(map[string]interface{}{
			"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			"userId": map[string]interface{}{
				"$in": []interface{}{utils.HexToMongoId(ctx, ctx.GetString("UserId")), nil},
			},
			"scope": map[string]interface{}{
				"$in": []constnotif.NotificationScope{
					constnotif.APP_WIDE_NOTIFICATION,
					constnotif.USER_NOTIFICATION,
					constnotif.WORKSPACE_NOTIFICATION,
				},
			},
		})
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return
		}
	}
	server_response.Response(ctx, http.StatusOK, "notifications fetched", true, notifications)
}

func DeleteNotifications(ctx *gin.Context) {
	var payload struct{
		Ids []string
	}
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{
			Err:        errors.New("pass in an array of notification ids to delete"),
			StatusCode: http.StatusBadRequest})
		return
	}
	success := notificationUseCases.DeleteNotificationsUseCase(ctx, payload.Ids)
	if !success{
		return
	}
	server_response.Response(ctx, http.StatusOK, "deleted successfully", success, nil)
}

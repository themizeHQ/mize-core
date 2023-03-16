package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	notificationModel "mize.app/app/notification/models"
	notification "mize.app/app/notification/repository"
	notificationUseCases "mize.app/app/notification/usecases"
	"mize.app/app_errors"
	"mize.app/server_response"
	"mize.app/utils"
)

func FetchUserNotifications(ctx *gin.Context) {
	notificationRepo := notification.GetNotificationRepo()
	var notifications *[]notificationModel.Notification
	var err error
	page, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
	if err != nil || page == 0 {
		page = 1
	}
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 15
	}
	skip := (page - 1) * limit
	// if ctx.GetString("Workspace") == "" {
	notifications, err = notificationRepo.FindMany(map[string]interface{}{
		"userId": map[string]interface{}{
			"$in": []interface{}{*utils.HexToMongoId(ctx, ctx.GetString("UserId")), nil},
		},
	}, options.Find().SetSort(map[string]interface{}{
		"updatedAt": -1,
	}), &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return
	}
	// } else {
	// notifications, err = notificationRepo.FindMany(map[string]interface{}{
	// 	"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	// 	"userId": map[string]interface{}{
	// 		"$in": []interface{}{utils.HexToMongoId(ctx, ctx.GetString("UserId")), nil},
	// 	},
	// 	"scope": map[string]interface{}{
	// 		"$in": []constnotif.NotificationScope{
	// 			constnotif.APP_WIDE_NOTIFICATION,
	// 			constnotif.USER_NOTIFICATION,
	// 			constnotif.WORKSPACE_NOTIFICATION,
	// 		},
	// 	},
	// }, &options.FindOptions{
	// 	Limit: &limit,
	// 	Skip:  &skip,
	// })
	// if err != nil {
	// 	app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
	// 	return
	// }
	// }
	server_response.Response(ctx, http.StatusOK, "notifications fetched", true, notifications)
}

func DeleteNotifications(ctx *gin.Context) {
	var payload struct {
		Ids []string
	}
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{
			Err:        errors.New("pass in an array of notification ids to delete"),
			StatusCode: http.StatusBadRequest})
		return
	}
	success := notificationUseCases.DeleteNotificationsUseCase(ctx, payload.Ids)
	if !success {
		return
	}
	server_response.Response(ctx, http.StatusOK, "deleted successfully", success, nil)
}

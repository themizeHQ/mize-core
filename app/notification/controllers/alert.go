package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mize.app/app/notification/models"
	"mize.app/app/notification/repository"
	alertUseCases "mize.app/app/notification/usecases/alert"
	"mize.app/app_errors"
	"mize.app/server_response"
	"mize.app/utils"
)

func SendAlert(ctx *gin.Context) {
	var payload models.Alert
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the valid json"), StatusCode: http.StatusBadRequest})
		return
	}
	success := alertUseCases.CreateNewAlertUseCase(ctx, payload)
	if !success {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "alert sent", true, nil)
}

func FetchAlerts(ctx *gin.Context) {
	alertRepo := repository.GetAlertRepo()
	startID := ctx.Query("id")
	limit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 15
	}
	alerts, err := alertRepo.FindManyStripped(map[string]interface{}{
		"usersId": []interface{}{utils.HexToMongoId(ctx, ctx.GetString("UserId"))},
		"_id": func() map[string]interface{} {
			if startID == "" {
				return map[string]interface{}{"$gt": primitive.NilObjectID}
			}
			return map[string]interface{}{"$lt": *utils.HexToMongoId(ctx, startID)}
		}(),
	}, options.Find().SetSort(map[string]interface{}{
		"updatedAt": -1,
	}), options.Find().SetSort(map[string]interface{}{
		"_id": -1,
	}), options.Find().SetProjection(map[string]interface{}{
		"message":     1,
		"importance":  1,
		"resourceUrl": 1,
		"resourceId":  1,
		"imageUrl":    1,
		"adminName":   1,
		"createdAt":   1,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return
	}
	server_response.Response(ctx, http.StatusOK, "alerts fetched", true, alerts)
}

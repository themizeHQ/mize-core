package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"mize.app/app/schedule/models"
	"mize.app/app/schedule/usecases"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func CreateSchedule(ctx *gin.Context) {
	var payload models.Schedule
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a valid schedule"), StatusCode: http.StatusBadRequest})
		return
	}
	usecases.CreateScheduleUseCase(ctx, &payload)
	server_response.Response(ctx, http.StatusCreated, "schedule created", true, nil)
}

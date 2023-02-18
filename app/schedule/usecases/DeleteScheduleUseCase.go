package usecases

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mize.app/app/schedule/repository"
	"mize.app/app_errors"
	"mize.app/logger"
	"mize.app/utils"
)

func DeleteScheduleUseCase(ctx *gin.Context, id string) error {
	scheduleRepo := repository.GetScheduleRepo()
	schedule, err := scheduleRepo.FindOneByFilter(map[string]interface{}{
		"_id":         id,
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	})
	if err != nil {
		logger.Error(errors.New("error fetching schedule to delete"), zap.Error(err))
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not remove schedule"), StatusCode: http.StatusInternalServerError})
		return err
	}
	if schedule == nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("schedule not found"), StatusCode: http.StatusNotFound})
		return err
	}
	success, err := scheduleRepo.DeleteOne(map[string]interface{}{
		"_id":         id,
		"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	})
	if err != nil {
		logger.Error(errors.New("error deleting schedule to delete"), zap.Error(err))
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not remove schedule"), StatusCode: http.StatusInternalServerError})
		return err
	}
	if !success {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("schedule not deleted"), StatusCode: http.StatusNotFound})
		return err
	}
	return nil
}

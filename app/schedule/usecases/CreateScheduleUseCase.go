package usecases

import (
	"time"

	"github.com/gin-gonic/gin"
	"mize.app/app/schedule/models"
	"mize.app/app/schedule/repository"
	"mize.app/schedule_manager"
	"mize.app/utils"
)

func CreateScheduleUseCase(ctx *gin.Context, payload *models.Schedule) {
	scheduleRepository := repository.GetScheduleRepo()
	payload.WorkspaceId = *utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
	scheduleRepository.CreateOne(*payload)
	now := time.Now().Unix()
	end := time.Now().Add(time.Hour * 1).Unix()
	for _, schedule := range payload.Events {
		if schedule.Time >= now && schedule.Time <= end {
			schedule_manager.Schedule(payload, schedule.Time, schedule_manager.Options{
				Url: schedule.Url,
			})
		}
	}
}

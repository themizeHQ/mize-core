package usecases

import (
	"sync"
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
	payload.CreatedBy = *utils.HexToMongoId(ctx, ctx.GetString("UserId"))
	scheduleRepository.CreateOne(*payload)
	now := time.Now().Unix()
	end := time.Now().Add(time.Hour * 1).Unix()
	wg := sync.WaitGroup{}
	for _, schedule := range payload.Events {
		if schedule.Time >= now && schedule.Time <= end {
			schedule_manager.Schedule(payload, schedule.Time, schedule_manager.Options{
				Url: schedule.Url,
			})
		}
		if payload.RemindByEmail {
			wg.Add(1)
			go schedule_manager.ScheduleEmail(payload, schedule, ctx.GetString("Workspace"))
			wg.Done()
		}
	}
	wg.Wait()
}

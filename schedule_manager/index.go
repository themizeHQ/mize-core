package schedule_manager

import (
	"context"
	"errors"

	// "fmt"
	"time"

	"github.com/procyon-projects/chrono"
	"go.uber.org/zap"
	scheduleRepository "mize.app/app/schedule/repository"
	"mize.app/logger"
)

// fetch schdules in db every to minitues and book them
func StartScheduleManager() {
	taskScheduler := chrono.NewDefaultTaskScheduler()
	taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		now := time.Now().Unix()
		end := time.Now().Add(time.Hour * 1).Unix()
		scheduleRepo := scheduleRepository.GetScheduleRepo()
		schedules, err := scheduleRepo.FindMany(map[string]interface{}{
			"time": map[string]interface{}{
				"$gte": now,
				"$lte": end,
			},
		})
		if err != nil {
			logger.Error(errors.New("could not fetch all user schedule on application start"), zap.Error(err))
			return
		}
		for _, schedule := range *schedules {
			if schedule.Time <= end || schedule.Time >= now {
				Schedule(&schedule, schedule.Time, Options{
					Url: schedule.Url,
				})
			}
			if schedule.RemindByEmail {
				ScheduleEmail(&schedule, schedule.WorkspaceId.Hex())
			}
		}
	}, 2*time.Hour, chrono.WithTime(time.Now()))
}

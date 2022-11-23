package schedule_manager

import (
	"context"
	"fmt"
	"time"

	"github.com/procyon-projects/chrono"
	scheduleRepository "mize.app/app/schedule/repository"
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
			fmt.Println("alert on sentry")
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

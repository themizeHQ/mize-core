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
			"events.time": map[string]interface{}{
				"$gte": now,
				"$lte": end,
			},
		})
		if err != nil {
			fmt.Println("alert on sentry")
			return
		}
		for _, schedule := range *schedules {
			for _, event := range schedule.Events {
				if event.Time <= end || event.Time >= now {
					Schedule(&schedule, event.Time, Options{
						Url: event.Url,
					})
				}
				if schedule.RemindByEmail {
					ScheduleEmail(&schedule, event, schedule.WorkspaceId.Hex())
				}
			}
		}
	}, 2*time.Hour, chrono.WithTime(time.Now()))
}

package schedule_manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/procyon-projects/chrono"
	"go.mongodb.org/mongo-driver/bson/primitive"

	notificationModels "mize.app/app/notification/models"
	alertRepoository "mize.app/app/notification/repository"
	scheduleModels "mize.app/app/schedule/models"
	teamModels "mize.app/app/teams/models"
	teamsRepo "mize.app/app/teams/repository"
	userRepo "mize.app/app/user/repository"
	notification_constants "mize.app/constants/notification"
	"mize.app/emails"
	"mize.app/realtime"
)

type Options struct {
	Url string
}

func Schedule(schedule *scheduleModels.Schedule, eventTime int64, opts Options) {
	taskScheduler := chrono.NewDefaultTaskScheduler()
	taskScheduler.Schedule(func(ctx context.Context) {

		var recipientIds []primitive.ObjectID
		for _, recipient := range schedule.Recipients {
			recipientIds = append(recipientIds, recipient.RecipientId)
		}
		alertRepo := alertRepoository.GetAlertRepo()
		_, err := alertRepo.CreateOne(notificationModels.Alert{
			Message:     schedule.Details,
			Importance:  schedule.Importance,
			ResourceUrl: opts.Url,
			UserIds:     recipientIds,
			AdminId:     schedule.CreatedBy,
			WorkspaceId: schedule.WorkspaceId,
		})
		if err != nil {
			fmt.Println("alert on sentry")
		}
		for _, recipient := range schedule.Recipients {
			realtime.CentrifugoController.Publish(recipient.RecipientId.Hex(), map[string]interface{}{
				"time":        time.Now(),
				"resourceUrl": opts.Url,
				"importance":  schedule.Importance,
				"message":     schedule.Details,
				"type":        notification_constants.SCHEDULE_REMINDER,
			})
		}

	}, chrono.WithTime(time.Unix(eventTime, 0)))
}

func ScheduleEmail(payload *scheduleModels.Schedule, schedule scheduleModels.Event, workspaceId string) {
	taskScheduler := chrono.NewDefaultTaskScheduler()
	taskScheduler.Schedule(func(ctx context.Context) {
		userRepo := userRepo.GetUserRepo()
		teamMemberRepo := teamsRepo.GetTeamMemberRepo()
		wg := sync.WaitGroup{}

		if payload.RemindByEmail {
			for _, recipient := range payload.Recipients {
				wg.Add(1)
				go func(rcp scheduleModels.Recipients) {
					defer func() {
						wg.Done()
					}()
					if rcp.Type == scheduleModels.UserRecipient {
						user, err := userRepo.FindById(rcp.RecipientId.Hex())
						if err != nil {
							return
						}
						emails.SendEmail(user.Email, payload.Name, "alert_reminder", map[string]interface{}{
							"TIME":    schedule.Time,
							"DETAILS": payload.Details,
						})
					} else if rcp.Type == scheduleModels.TeamRecipient {
						members, err := teamMemberRepo.FindMany(map[string]interface{}{
							"workspaceId": workspaceId,
						})
						if err != nil {
							return
						}

						for _, member := range *members {
							wg.Add(1)
							go func(mem teamModels.TeamMembers) {
								defer func() {
									wg.Done()
								}()
								user, err := userRepo.FindById(mem.UserId.String())
								if err != nil {
									return
								}
								emails.SendEmail(user.Email, payload.Name, "alert_reminder", map[string]interface{}{
									"TIME":    time.Unix(schedule.Time, 0),
									"DETAILS": payload.Details,
								})
							}(member)
						}
					}
				}(recipient)
			}
			wg.Wait()
		}
	}, chrono.WithTime(time.Unix(schedule.Time-1800, 0)))

}

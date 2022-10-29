package schedule_manager

import (
	"context"
	"fmt"
	"time"

	"github.com/procyon-projects/chrono"
	"go.mongodb.org/mongo-driver/bson/primitive"

	notificationModels "mize.app/app/notification/models"
	alertRepoository "mize.app/app/notification/repository"
	scheduleModels "mize.app/app/schedule/models"
	notification_constants "mize.app/constants/notification"
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

package schedule_manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/procyon-projects/chrono"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	notificationModels "mize.app/app/notification/models"
	notificationRepoository "mize.app/app/notification/repository"
	scheduleModels "mize.app/app/schedule/models"
	teamModels "mize.app/app/teams/models"
	teamsRepo "mize.app/app/teams/repository"
	userRepo "mize.app/app/user/repository"
	notification_constants "mize.app/constants/notification"
	"mize.app/emails"
	"mize.app/logger"
	"mize.app/realtime"
	"mize.app/sms"
)

type Options struct {
	Url *string
}

func Schedule(schedule *scheduleModels.Schedule, eventTime int64, opts Options) {
	taskScheduler := chrono.NewDefaultTaskScheduler()
	taskScheduler.Schedule(func(ctx context.Context) {
		var recipientIds []primitive.ObjectID
		if schedule.Recipients != nil {
			for _, recipient := range *schedule.Recipients {
				recipientIds = append(recipientIds, recipient.RecipientId)
			}
		}
		var err error
		if schedule.WorkspaceId != nil {
			alertRepo := notificationRepoository.GetAlertRepo()
			_, err = alertRepo.CreateOne(notificationModels.Alert{
				Message:     schedule.Details,
				Importance:  schedule.Importance,
				ResourceUrl: opts.Url,
				UserIds:     recipientIds,
				AdminId:     schedule.CreatedBy,
				WorkspaceId: *schedule.WorkspaceId,
			})

		} else {
			notifiRepo := notificationRepoository.GetNotificationRepo()
			createdAtPointer := schedule.CreatedBy
			_, err = notifiRepo.CreateOne(notificationModels.Notification{
				Message:     schedule.Details,
				Type:        notification_constants.SCHEDULE_REMINDER,
				Importance:  schedule.Importance,
				Scope:       notification_constants.USER_NOTIFICATION,
				ExternalUrl: schedule.Url,
				UserId:      &createdAtPointer,
			})
		}
		if err != nil {
			logger.Error(errors.New("schedule error - could not create notification"), zap.Error(err))
		}
		if schedule.Recipients != nil {
			for _, recipient := range *schedule.Recipients {
				realtime.CentrifugoController.Publish(recipient.RecipientId.Hex(), map[string]interface{}{
					"time":        time.Now(),
					"resourceUrl": opts.Url,
					"importance":  schedule.Importance,
					"message":     schedule.Details,
					"type":        notification_constants.SCHEDULE_REMINDER,
				})
			}
		}

	}, chrono.WithTime(time.Unix(eventTime, 0)))
}

func ScheduleEmail(payload *scheduleModels.Schedule, workspaceId string) {
	taskScheduler := chrono.NewDefaultTaskScheduler()
	taskScheduler.Schedule(func(ctx context.Context) {
		userRepo := userRepo.GetUserRepo()
		teamMemberRepo := teamsRepo.GetTeamMemberRepo()
		wg := sync.WaitGroup{}

		if payload.Recipients != nil {
			formatedTime := time.Unix(payload.Time, 0).Local().Format(time.UnixDate)
			for _, recipient := range *payload.Recipients {
				wg.Add(1)
				go func(rcp scheduleModels.Recipients) {
					defer func() {
						wg.Done()
					}()
					if rcp.Type == scheduleModels.UserRecipient {
						user, err := userRepo.FindById(rcp.RecipientId.Hex())
						if err != nil {
							logger.Error(errors.New("schedule error - could not find user"), zap.Error(err))
							return
						}
						if user == nil {
							return
						}
						emails.SendEmail(user.Email, fmt.Sprintf("Schedule Reminder from %s", payload.From), "schedule_reminder", map[string]interface{}{
							"TIME":     formatedTime,
							"DETAILS":  payload.Details,
							"NAME":     payload.Name,
							"LOCATION": payload.Location,
							"FROM":     payload.From,
							"URL":      payload.Url,
						})
					} else if rcp.Type == scheduleModels.TeamRecipient {
						members, err := teamMemberRepo.FindMany(map[string]interface{}{
							"workspaceId": workspaceId,
							"teamId":      rcp.RecipientId,
						})
						if err != nil {
							logger.Error(errors.New("schedule error - could not find team members"), zap.Error(err))
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
									logger.Error(errors.New("schedule error -could not create notification"), zap.Error(err))
									return
								}
								emails.SendEmail(user.Email, fmt.Sprintf("Schedule Reminder from %s", payload.From), "schedule_reminder", map[string]interface{}{
									"TIME":     formatedTime,
									"DETAILS":  payload.Details,
									"NAME":     payload.Name,
									"LOCATION": payload.Location,
									"FROM":     payload.From,
									"URL":      payload.Url,
								})
							}(member)
						}
					}
				}(recipient)
			}
			wg.Wait()
		}
		// remind 30 mins before time
	}, chrono.WithTime(time.Unix(payload.Time-1800, 0)))

}

func ScheduleSMS(payload *scheduleModels.Schedule, workspaceId string) {
	taskScheduler := chrono.NewDefaultTaskScheduler()
	taskScheduler.Schedule(func(ctx context.Context) {
		userRepo := userRepo.GetUserRepo()
		teamMemberRepo := teamsRepo.GetTeamMemberRepo()
		wg := sync.WaitGroup{}

		if payload.Recipients != nil {
			formatedTime := time.Unix(payload.Time, 0).Local().Format(time.UnixDate)
			for _, recipient := range *payload.Recipients {
				wg.Add(1)
				go func(rcp scheduleModels.Recipients) {
					defer func() {
						wg.Done()
					}()
					if rcp.Type == scheduleModels.UserRecipient {
						user, err := userRepo.FindOneByFilter(map[string]interface{}{
							"_id":   rcp.RecipientId.Hex(),
							"phone": map[string]interface{}{"$ne": nil},
						}, options.FindOne().SetProjection(map[string]interface{}{
							"firstName": 1,
							"phone":     1,
						}))
						if err != nil || user == nil {
							logger.Error(errors.New("schedule error - could not find user"), zap.Error(err))
							return
						}
						sms.SmsService.SendSms(*user.Phone, fmt.Sprintf("Hi %s, you've got an event '%s' scheduled for %s.\nVenue is %s.", user.FirstName, payload.Name, formatedTime, payload.Location))
					} else if rcp.Type == scheduleModels.TeamRecipient {
						members, err := teamMemberRepo.FindMany(map[string]interface{}{
							"workspaceId": workspaceId,
							"teamId":      rcp.RecipientId,
						})
						if err != nil {
							logger.Error(errors.New("schedule error - could not find team members"), zap.Error(err))
							return
						}

						for _, member := range *members {
							wg.Add(1)
							go func(mem teamModels.TeamMembers) {
								defer func() {
									wg.Done()
								}()
								user, err := userRepo.FindOneByFilter(map[string]interface{}{
									"_id":   rcp.RecipientId.Hex(),
									"phone": map[string]interface{}{"$ne": nil},
								}, options.FindOne().SetProjection(map[string]interface{}{
									"firstName": 1,
									"phone":     1,
								}))
								if err != nil || user == nil {
									logger.Error(errors.New("schedule error - could not find team member "), zap.Error(err))
									return
								}
								sms.SmsService.SendSms(*user.Phone, fmt.Sprintf("Hi %s, you've got an event '%s' scheduled for %s at %s.", user.FirstName, payload.Name, formatedTime, payload.Location))
							}(member)
						}
					}
				}(recipient)
			}
			wg.Wait()
		}
		// remind 30 mins before time
	}, chrono.WithTime(time.Unix(payload.Time-1800, 0)))
	logger.Info("schedule reminder for sms set", zap.String("scheduleID", payload.Id.Hex()))
}

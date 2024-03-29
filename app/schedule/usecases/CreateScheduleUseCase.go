package usecases

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mize.app/app/schedule/models"
	"mize.app/app/schedule/repository"
	teamRepository "mize.app/app/teams/repository"
	userRepository "mize.app/app/user/repository"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	workspace_member_constants "mize.app/constants/workspace"
	"mize.app/logger"
	"mize.app/schedule_manager"
	"mize.app/utils"
)

func CreateScheduleUseCase(ctx *gin.Context, payload *models.Schedule) {
	scheduleRepository := repository.GetScheduleRepo()
	payload.WorkspaceId = utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
	payload.CreatedBy = *utils.HexToMongoId(ctx, ctx.GetString("UserId"))
	payload.From = fmt.Sprintf("%s %s", ctx.GetString("Firstname"), ctx.GetString("Lastname"))
	if len(*payload.Recipients) != 0 {
		workspaceRepo := workspaceRepo.GetWorkspaceMember()
		member, err := workspaceRepo.FindOneByFilter(map[string]interface{}{
			"admin":  true,
			"userId": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		}, options.FindOne().SetProjection(map[string]interface{}{
			"adminAccess": 1,
		}))
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("an error occured while validating user"), StatusCode: http.StatusInternalServerError})
			return
		}
		access := member.HasAccess([]workspace_member_constants.AdminAccessType{workspace_member_constants.AdminAccess.SCHEDULE_ACCESS, workspace_member_constants.AdminAccess.FULL_ACCESS})
		if !access {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you do not have authorization to create schedules for other members"), StatusCode: http.StatusInternalServerError})
			return
		}
	}
	wg := sync.WaitGroup{}
	userRepo := userRepository.GetUserRepo()
	teamRepo := teamRepository.GetTeamRepo()
	recipients := *payload.Recipients
	for index, recipient := range recipients {
		wg.Add(1)
		go func(r models.Recipients, i int) {
			defer func() {
				wg.Done()
			}()
			if r.Type == models.UserRecipient {
				user, err := userRepo.FindById(r.RecipientId.Hex(), options.FindOne().SetProjection(map[string]interface{}{
					"firstName": 1,
					"lastName":  1,
				}))
				if err != nil {
					logger.Error(errors.New("an error occured while fetching user after schedule was created. name not populated"))
					return
				}
				r.Name = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
				recipients[i] = r
			} else if r.Type == models.TeamRecipient {
				team, err := teamRepo.FindOneByFilter(map[string]interface{}{
					"_id":         r.RecipientId.Hex(),
					"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
				}, options.FindOne().SetProjection(map[string]interface{}{
					"name": 1,
				}))
				if err != nil {
					logger.Error(errors.New("an error occured while fetching team after schedule was created. name not populated"))
					return
				}
				r.Name = team.Name
				recipients[i] = r
			} else {
				logger.Error(errors.New("illegal recipient type passed for schedule"))
				return
			}
		}(recipient, index)
	}
	wg.Wait()
	payload.Recipients = &recipients
	scheduleRepository.CreateOne(*payload)
	now := time.Now().Unix()
	end := time.Now().Add(time.Hour * 1).Unix()
	wg = sync.WaitGroup{}
	if payload.Time >= now && payload.Time <= end {
		schedule_manager.Schedule(payload, payload.Time, schedule_manager.Options{
			Url: payload.Url,
		})
	}
	if payload.RemindByEmail {
		wg.Add(1)
		go schedule_manager.ScheduleEmail(payload, ctx.GetString("Workspace"))
		wg.Done()
	}
	if payload.RemindBySMS {
		wg.Add(1)
		go schedule_manager.ScheduleSMS(payload, ctx.GetString("Workspace"))
		wg.Done()
	}
	wg.Wait()
}

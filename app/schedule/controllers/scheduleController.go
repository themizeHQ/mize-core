package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mize.app/app/schedule/models"
	schedulesRepo "mize.app/app/schedule/repository"
	"mize.app/app/schedule/types"
	"mize.app/app/schedule/usecases"
	"mize.app/utils"

	teamsRepo "mize.app/app/teams/repository"
	"mize.app/app_errors"
	"mize.app/server_response"
)

func CreateSchedule(ctx *gin.Context) {
	var payload types.SchedulePayload
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a valid schedule"), StatusCode: http.StatusBadRequest})
		return
	}
	payload.SetScheduleTime()
	usecases.CreateScheduleUseCase(ctx, &payload.Payload)
	server_response.Response(ctx, http.StatusCreated, "schedule created", true, nil)
}

func FetchUserSchedule(ctx *gin.Context) {
	scheduleRepository := schedulesRepo.GetScheduleRepo()
	teamRepository := teamsRepo.GetTeamMemberRepo()
	teamMembers, err := teamRepository.FindMany(map[string]interface{}{}, options.Find().SetProjection(map[string]interface{}{
		"_id": 1,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not fetch schedule"), StatusCode: http.StatusBadRequest})
		return
	}
	var recipientFilter []map[string]interface{}
	for _, member := range *teamMembers {
		recipientFilter = append(recipientFilter, map[string]interface{}{
			"recipientId": member.Id.Hex(),
			"type":        models.TeamRecipient,
			"name":        member.TeamName,
		})
	}
	recipientFilter = append(recipientFilter, map[string]interface{}{
		"recipientId": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"type":        models.UserRecipient,
		"name":        fmt.Sprintf("%s %s", ctx.GetString("Firstname"), ctx.GetString("Lastname")),
	})
	events, err := scheduleRepository.FindManyStripped(map[string]interface{}{
		"recipient.recipientId": map[string]interface{}{
			"$in": recipientFilter,
		},
	}, options.Find().SetProjection(map[string]interface{}{
		"name":     1,
		"details":  1,
		"location": 1,
		"time":     1,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not fetch schedule"), StatusCode: http.StatusBadRequest})
		return
	}
	fmt.Println(events)
	server_response.Response(ctx, http.StatusCreated, "schedule fetched", true, events)
}

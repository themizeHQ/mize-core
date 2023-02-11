package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	teamMemberRepository := teamsRepo.GetTeamMemberRepo()
	teamMembers, err := teamMemberRepository.FindMany(map[string]interface{}{}, options.Find().SetProjection(map[string]interface{}{
		"userId": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not fetch schedule"), StatusCode: http.StatusBadRequest})
		return
	}
	var recipientIdFilter []primitive.ObjectID
	for _, member := range *teamMembers {
		recipientIdFilter = append(recipientIdFilter, member.Id)
	}
	recipientIdFilter = append(recipientIdFilter, *utils.HexToMongoId(ctx, ctx.GetString("UserId")))
	events, err := scheduleRepository.FindManyStripped(map[string]interface{}{
		"recipient": map[string]interface{}{
			"$elemMatch": map[string]interface{}{
				"recipientId": map[string]interface{}{
					"$in": recipientIdFilter,
				},
			},
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
	server_response.Response(ctx, http.StatusOK, "schedule fetched", true, events)
}

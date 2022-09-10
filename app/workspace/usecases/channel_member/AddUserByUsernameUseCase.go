package channel_member

import (
	"errors"
	"fmt"
	"net/http"

	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/workspace/models"
	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	channel_constants "mize.app/constants/channel"
	"mize.app/utils"
)

func AddUserByUsernameUseCase(ctx *gin.Context, username string, channel_id string) bool {
	var wg sync.WaitGroup
	channelMemberRepo := repository.GetChannelMemberRepo()

	chan1 := make(chan app_errors.RequestError)
	wg.Add(1)
	go func(ch chan app_errors.RequestError) {
		defer func() {
			wg.Done()
		}()
		admin, err := channelMemberRepo.FindOneByFilter(map[string]interface{}{
			"channelId": utils.HexToMongoId(ctx, channel_id),
			"userId":    utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		},
			options.FindOne().SetProjection(map[string]int{
				"admin":       1,
				"adminAccess": 1,
			}),
		)
		if err != nil {
			ch <- app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError}
			return
		}
		if admin == nil {
			ch <- app_errors.RequestError{Err: errors.New("you are not a member of this channel"), StatusCode: http.StatusInternalServerError}
			return
		}
		if !admin.Admin {
			ch <- app_errors.RequestError{Err: errors.New("only admins can add users to channels"), StatusCode: http.StatusUnauthorized}
			return
		}
		has_access := admin.HasAccess(admin.AdminAccess, []channel_constants.ChannelAdminAccess{channel_constants.CHANNEL_MEMBERSHIP_ACCESS, channel_constants.CHANNEL_FULL_ACCESS})
		if !has_access {
			ch <- app_errors.RequestError{Err: errors.New("you do not have this permission"), StatusCode: http.StatusUnauthorized}
			return
		}
		ch <- app_errors.RequestError{}
	}(chan1)

	chan2 := make(chan app_errors.RequestError)
	wg.Add(1)
	go func(ch chan app_errors.RequestError) {
		defer func() {
			wg.Done()
		}()
		exists, err := channelMemberRepo.CountDocs(map[string]interface{}{
			"userName":  username,
			"channelId": *utils.HexToMongoId(ctx, channel_id),
		})
		if err != nil {
			ch <- app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError}
			return
		}
		if exists > 0 {
			ch <- app_errors.RequestError{Err: fmt.Errorf("%s is already a member of this channel", username), StatusCode: http.StatusConflict}
			return
		}
		ch <- app_errors.RequestError{}
	}(chan2)

	chan3 := make(chan app_errors.RequestError)
	channelName := make(chan string)
	wg.Add(1)
	go func(ch chan app_errors.RequestError, name chan string) {
		defer func() {
			wg.Done()
			fmt.Println(<-ch)
		}()
		channelRepo := repository.GetChannelRepo()
		channel, err := channelRepo.FindOneByFilter(map[string]interface{}{
			"_id":         channel_id,
			"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		}, options.FindOne().SetProjection(map[string]int{
			"name": 1,
		}))

		if err != nil {
			ch <- app_errors.RequestError{Err: errors.New("something went wrong while adding user"), StatusCode: http.StatusInternalServerError}
			name <- ""
			return
		}
		if channel == nil {
			ch <- app_errors.RequestError{Err: errors.New("channel does not exist"), StatusCode: http.StatusNotFound}
			name <- ""
			return
		}
		ch <- app_errors.RequestError{}
		name <- channel.Name
	}(chan3, channelName)

	chan4 := make(chan app_errors.RequestError)
	userId := make(chan primitive.ObjectID)
	wg.Add(1)
	go func(ch chan app_errors.RequestError, id chan primitive.ObjectID) {
		defer func() {
			wg.Done()
		}()
		workspaceRepository := repository.GetWorkspaceMember()
		user, err := workspaceRepository.FindOneByFilter(map[string]interface{}{
			"userName": username,
		}, options.FindOne().SetProjection(map[string]int{
			"_id": 1,
		}))
		if err != nil {
			ch <- app_errors.RequestError{Err: errors.New("something went wrong while adding user"), StatusCode: http.StatusInternalServerError}
			id <- primitive.NilObjectID
			return
		}
		if user == nil {
			ch <- app_errors.RequestError{Err: fmt.Errorf("%s is not a member of this workspace", username), StatusCode: http.StatusNotFound}
			id <- primitive.NilObjectID
			return
		}
		ch <- app_errors.RequestError{}
		id <- user.Id
	}(chan4, userId)

	err1 := <-chan1
	if err1.Err != nil {
		app_errors.ErrorHandler(ctx, err1)
		return false
	}

	err2 := <-chan2
	if err2.Err != nil {
		app_errors.ErrorHandler(ctx, err2)
		return false
	}

	err3 := <-chan3
	if err3.Err != nil {
		app_errors.ErrorHandler(ctx, err3)
		return false
	}

	err4 := <-chan4
	if err4.Err != nil {
		app_errors.ErrorHandler(ctx, err4)
		return false
	}
	member := models.ChannelMember{
		ChannelId:   *utils.HexToMongoId(ctx, channel_id),
		WorkspaceId: *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		Username:    username,
		UserId:      <-userId,
		LastSent:    primitive.NewDateTimeFromTime(time.Now()),
		Admin:       false,
		ChannelName: <-channelName,
		AdminAccess: []channel_constants.ChannelAdminAccess{},
		JoinDate:    primitive.NewDateTimeFromTime(time.Now()),
	}
	wg.Wait()

	if err := member.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return false
	}
	_, err := channelMemberRepo.CreateOne(member)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return false
	}
	return true
}

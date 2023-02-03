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

	userModels "mize.app/app/user/models"
	userRepo "mize.app/app/user/repository"
	"mize.app/app/workspace/models"
	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	channel_constants "mize.app/constants/channel"
	"mize.app/utils"
)

func AdminAddUserUseCase(ctx *gin.Context, id string, channel_id string, addBy string) bool {
	var wg sync.WaitGroup
	channelMemberRepo := repository.GetChannelMemberRepo()

	chan1 := make(chan app_errors.RequestError)
	channelName := make(chan *string)
	wg.Add(1)
	go func(ch chan app_errors.RequestError, name chan *string) {
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
			name <- nil
			return
		}
		if admin == nil {
			ch <- app_errors.RequestError{Err: errors.New("you are not a member of this channel"), StatusCode: http.StatusInternalServerError}
			name <- nil
			return
		}
		if !admin.Admin {
			ch <- app_errors.RequestError{Err: errors.New("only admins can add users to channels"), StatusCode: http.StatusUnauthorized}
			name <- nil
			return
		}
		has_access := admin.HasAccess([]channel_constants.ChannelAdminAccess{channel_constants.CHANNEL_MEMBERSHIP_ACCESS, channel_constants.CHANNEL_FULL_ACCESS})
		if !has_access {
			ch <- app_errors.RequestError{Err: errors.New("you do not have this permission"), StatusCode: http.StatusUnauthorized}
			name <- nil
			return
		}
		ch <- app_errors.RequestError{}
		name <- &admin.ChannelName
	}(chan1, channelName)

	chan3 := make(chan app_errors.RequestError)
	userChan := make(chan *userModels.User)
	wg.Add(1)
	go func(ch chan app_errors.RequestError, u chan *userModels.User) {
		defer func() {
			wg.Done()
		}()
		userRepo := userRepo.GetUserRepo()
		var user *userModels.User
		var err error
		if addBy == "username" {
			user, err = userRepo.FindOneByFilter(map[string]interface{}{
				"userName": id,
			}, options.FindOne().SetProjection(map[string]int{
				"_id":      1,
				"userName": 1,
			}))
		} else if addBy == "email" {
			user, err = userRepo.FindOneByFilter(map[string]interface{}{
				"email": id,
			}, options.FindOne().SetProjection(map[string]int{
				"_id":      1,
				"userName": 1,
			}))
		} else if addBy == "phone" {
			user, err = userRepo.FindOneByFilter(map[string]interface{}{
				"phone": id,
			}, options.FindOne().SetProjection(map[string]int{
				"_id":      1,
				"userName": 1,
			}))
		} else {
			ch <- app_errors.RequestError{Err: errors.New("pass in valid add method"), StatusCode: http.StatusConflict}
			u <- nil
			return
		}
		if err != nil {
			ch <- app_errors.RequestError{Err: errors.New("something went wrong while adding user"), StatusCode: http.StatusInternalServerError}
			u <- nil
			return
		}
		if user == nil {
			ch <- app_errors.RequestError{Err: fmt.Errorf("%s is not registered to an account", id), StatusCode: http.StatusNotFound}
			u <- nil
			return
		}
		workspaceRepository := repository.GetWorkspaceMember()
		var member int64
		member, err = workspaceRepository.CountDocs(map[string]interface{}{
			"userName":    user.UserName,
			"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		})
		if err != nil {
			ch <- app_errors.RequestError{Err: errors.New("something went wrong while adding user"), StatusCode: http.StatusInternalServerError}
			u <- nil
			return
		}
		if member == 0 {
			ch <- app_errors.RequestError{Err: fmt.Errorf("%s is not a member of this workspace", id), StatusCode: http.StatusNotFound}
			u <- nil
			return
		}
		ch <- app_errors.RequestError{}
		u <- user
	}(chan3, userChan)

	chan2 := make(chan app_errors.RequestError)
	wg.Add(1)
	go func(ch chan app_errors.RequestError) {
		defer func() {
			wg.Done()
		}()

		userRepo := userRepo.GetUserRepo()
		var user *userModels.User
		var err error
		if addBy == "username" {
			user, err = userRepo.FindOneByFilter(map[string]interface{}{
				"userName": id,
			}, options.FindOne().SetProjection(map[string]int{
				"_id":      1,
				"userName": 1,
			}))
		} else if addBy == "email" {
			user, err = userRepo.FindOneByFilter(map[string]interface{}{
				"email": id,
			}, options.FindOne().SetProjection(map[string]int{
				"_id":      1,
				"userName": 1,
			}))
		} else if addBy == "phone" {
			user, err = userRepo.FindOneByFilter(map[string]interface{}{
				"phone": id,
			}, options.FindOne().SetProjection(map[string]int{
				"_id":      1,
				"userName": 1,
			}))
		} else {
			ch <- app_errors.RequestError{Err: errors.New("pass in valid add method"), StatusCode: http.StatusConflict}
			return
		}
		if err != nil {
			ch <- app_errors.RequestError{Err: errors.New("something went wrong while adding user"), StatusCode: http.StatusInternalServerError}
			return
		}
		if user == nil {
			ch <- app_errors.RequestError{Err: fmt.Errorf("%s is not registered to an account", id), StatusCode: http.StatusNotFound}
			return
		}

		exists, err := channelMemberRepo.CountDocs(map[string]interface{}{
			"userId":    user.Id,
			"channelId": *utils.HexToMongoId(ctx, channel_id),
		})
		if err != nil {
			ch <- app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError}
			return
		}
		if exists != 0 {
			ch <- app_errors.RequestError{Err: fmt.Errorf("%s is already a member of this channel", id), StatusCode: http.StatusConflict}
			return
		}
		ch <- app_errors.RequestError{}
	}(chan2)

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

	user := <-userChan
	member := models.ChannelMember{
		ChannelId:   *utils.HexToMongoId(ctx, channel_id),
		WorkspaceId: *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		Username:    user.UserName,
		UserId:      user.Id,
		LastSent:    primitive.NewDateTimeFromTime(time.Now()),
		Admin:       false,
		ChannelName: *<-channelName,
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

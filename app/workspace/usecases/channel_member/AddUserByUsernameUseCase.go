package channel_member

import (
	"errors"
	"fmt"
	"net/http"
	// "sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	userRepo "mize.app/app/user/repository"
	"mize.app/app/workspace/models"
	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	channel_constants "mize.app/constants/channel"
	"mize.app/utils"
)

func AddUserByUsernameUseCase(ctx *gin.Context, username string, channel_id string) bool {
	// var wg sync.WaitGroup
	channelMemberRepo := repository.GetChannelMemberRepo()

	chan1 := make(chan app_errors.RequestError)
	// wg.Add(1)
	go func(ch chan app_errors.RequestError) {
		// defer func() {
		// 	wg.Done()
		// }()
		admin, err := channelMemberRepo.FindById(ctx.GetString("UserId"),
			options.FindOne().SetProjection(map[string]int{
				"admin":       1,
				"adminAccess": 1,
			}),
		)
		if err != nil {
			fmt.Println(err)
			ch <- app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError}
			return
		}
		if !admin.Admin {
			err = errors.New("only admins can add users to channels")
			fmt.Println("2 err")
			ch <- app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized}
			return
		}
		has_access := admin.HasAccess(admin.AdminAccess, []channel_constants.ChannelAdminAccess{channel_constants.CHANNEL_MEMBERSHIP_ACCESS})
		if !has_access {
			err = errors.New("you do not have this permission")
			fmt.Println("3 err")
			ch <- app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized}
			return
		}
		fmt.Println("no error")
		ch <- app_errors.RequestError{}
		fmt.Println("no error")
	}(chan1)

	chan2 := make(chan app_errors.RequestError)
	// wg.Add(1)
	go func(ch chan app_errors.RequestError) {
		// defer func() {
		// 	wg.Done()
		// }()
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
	// wg.Add(1)
	go func(ch chan app_errors.RequestError, name chan string) {
		// defer func() {
		// 	wg.Done()
		// }()
		channelRepo := repository.GetChannelRepo()
		channel, err := channelRepo.FindOneByFilter(map[string]interface{}{
			"_id":         channel_id,
			"workspaceId": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		}, options.FindOne().SetProjection(map[string]int{
			"name": 1,
		}))
		if err != nil {
			ch <- app_errors.RequestError{Err: errors.New("something went wrong while adding user"), StatusCode: http.StatusInternalServerError}
			return
		}
		if channel == nil {
			ch <- app_errors.RequestError{Err: errors.New("channel does not exist"), StatusCode: http.StatusNotFound}
			return
		}
		ch <- app_errors.RequestError{}
		name <- channel.Name
	}(chan3, channelName)

	chan4 := make(chan app_errors.RequestError)
	userId := make(chan primitive.ObjectID)
	// wg.Add(1)
	go func(ch chan app_errors.RequestError, id chan primitive.ObjectID) {
		// defer func() {
		// 	wg.Done()
		// }()
		userRepository := userRepo.GetUserRepo()
		user, err := userRepository.FindOneByFilter(map[string]interface{}{
			"userName": username,
		}, options.FindOne().SetProjection(map[string]int{
			"_id": 1,
		}))
		if err != nil {
			ch <- app_errors.RequestError{Err: errors.New("something went wrong while adding user"), StatusCode: http.StatusInternalServerError}
			return
		}
		if user == nil {
			ch <- app_errors.RequestError{Err: fmt.Errorf("user %s not found", username), StatusCode: http.StatusNotFound}
			return
		}
		ch <- app_errors.RequestError{}
		id <- user.Id
	}(chan4, userId)

	fmt.Println("1")
	if (<-chan1).Err != nil {
		app_errors.ErrorHandler(ctx, <-chan1)
		return false
	}

	fmt.Println("2")
	if (<-chan2).Err != nil {
		app_errors.ErrorHandler(ctx, <-chan2)
		return false
	}

	if (<-chan3).Err != nil {
		app_errors.ErrorHandler(ctx, <-chan3)
		return false
	}

	if (<-chan4).Err != nil {
		app_errors.ErrorHandler(ctx, <-chan4)
		return false
	}
	fmt.Println("3")
	// wg.Wait()

	member := models.ChannelMember{
		ChannelId:   *utils.HexToMongoId(ctx, channel_id),
		WorkspaceId: *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		Username:    username,
		UserId:      primitive.NilObjectID,
		LastSent:    primitive.NewDateTimeFromTime(time.Now()),
		Admin:       false,
		ChannelName: "",
		AdminAccess: []channel_constants.ChannelAdminAccess{},
		JoinDate:    primitive.NewDateTimeFromTime(time.Now()),
	}
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

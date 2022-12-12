package usecases

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"mize.app/app/user/models"
	"mize.app/app/user/repository"
	"mize.app/app_errors"
	user_constants "mize.app/constants/user"
	"mize.app/repository/database/redis"
)

func CacheUserUseCase(ctx *gin.Context, payload models.User) error {
	var userRepoInstance = repository.GetUserRepo()
	payload.Status = user_constants.AVAILABLE
	if payload.Language == "" {
		payload.Language = "english"
	}
	payload.Discoverability = []user_constants.UserDiscoverability{user_constants.DISCOVERABILITY_EMAIL, user_constants.DISCOVERABILITY_USERNAME, user_constants.DISCOVERABILITY_PHONE}
	payload.RunHooks()
	var wg sync.WaitGroup
	errChan := make(chan error)
	wg.Add(1)
	go func(e chan error) {
		defer func() {
			wg.Done()
		}()
		emailExists, err := userRepoInstance.CountDocs(map[string]interface{}{"email": payload.Email})
		if err != nil {
			e <- err
			return
		}
		if emailExists != 0 {
			e <- errors.New("user with email already exists")
			return
		}
		e <- nil
	}(errChan)
	wg.Add(1)
	go func(e chan error) {
		defer func() {
			wg.Done()
		}()
		usernameExists, err := userRepoInstance.CountDocs(map[string]interface{}{"userName": payload.UserName})
		if err != nil {
			e <- err
			return
		}
		if usernameExists != 0 {
			e <- errors.New("user with username already exists")
			return
		}
		e <- nil
	}(errChan)
	err := <-errChan
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
		return err
	}
	result := redis.RedisRepo.CreateEntry(ctx, fmt.Sprintf("%s-user", payload.Email), payload, 20*time.Minute)
	if !result {
		err := errors.New("something went wrong while creating user")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return err
	}
	return nil
}

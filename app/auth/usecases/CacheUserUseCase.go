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
	"mize.app/repository/database/redis"
)

func CacheUserUseCase(ctx *gin.Context, payload models.User) error {
	var userRepoInstance = repository.GetUserRepo()
	payload.RunHooks()
	var wg sync.WaitGroup
	emailExistsErrChan := make(chan error)
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
		if emailExists == 1 {
			e <- errors.New("user with email already exists")
			return
		}
		e <- nil
	}(emailExistsErrChan)
	usernameExistsErrChan := make(chan error)
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
		if usernameExists == 1 {
			e <- errors.New("user with username already exists")
			return
		}
		e <- nil
	}(usernameExistsErrChan)
	err := <-usernameExistsErrChan
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
		return err
	}
	err = <-emailExistsErrChan
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

package application

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	application "mize.app/app/application/models"
	appRepo "mize.app/app/application/repository"
	userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
)

var accessibleUserData = []string{"firstName", "lastName", "userName", "email", "region"}

func CreateAppUseCase(ctx *gin.Context, payload application.Application) (*mongo.InsertOneResult, error) {
	appNameExists := appRepo.AppRepository.CountDocs(ctx, map[string]interface{}{"name": payload.Name})
	user := userRepo.UserRepository.FindOneByFilter(ctx, map[string]interface{}{"_id": ctx.GetString("UserId")})
	if user == nil {
		err := errors.New("user does not exist")
		app_errors.ErrorHandler(ctx, err, http.StatusNotFound)
		return nil, err
	}
	payload.CreatedBy = user.Id
	payload.Email = user.Email
	payload.Approved = false
	payload.Active = false

	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return nil, err
	}
	if appNameExists != 0 {
		err := fmt.Errorf("name %s has already been taken by another application", payload.Name)
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return nil, err
	}
	for _, data := range payload.RequiredData {
		var invalid bool = true
		for _, d := range accessibleUserData {
			if d == data {
				invalid = false
			}
		}
		if invalid {
			err := fmt.Errorf("data field %s is not a valid user field to access", data)
			app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
			return nil, err
		}
	}
	response := appRepo.AppRepository.CreateOne(ctx, &payload)
	return &response, nil
}

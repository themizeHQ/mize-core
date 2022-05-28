package application

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	application "mize.app/app/application/models"
	appRepo "mize.app/app/application/repository"
	"mize.app/app_errors"
)

func CreateAppUseCase(ctx *gin.Context, payload application.Application) (*mongo.InsertOneResult, error) {
	appNameExists := appRepo.AppRepository.CountDocs(ctx, map[string]interface{}{"name": payload.Name})
	if appNameExists != 0 {
		err := fmt.Errorf("name %s has already been taken by another application", payload.Name)
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return nil, err
	}
	response := appRepo.AppRepository.CreateOne(ctx, &payload)
	return &response, nil
}

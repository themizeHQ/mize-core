package usecases

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/application/models"
	"mize.app/app/application/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

var accessibleUserData = []string{"firstName", "lastName", "userName", "email", "region"}

func CreateAppUseCase(ctx *gin.Context, payload models.Application) error {
	var appRepoInstance = repository.GetAppRepo()
	appNameExists, err := appRepoInstance.CountDocs(map[string]interface{}{"name": payload.Name})
	if err != nil {
		return err
	}
	if appNameExists != 0 {
		err = fmt.Errorf("name %s has already been taken by another application", payload.Name)
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
		return err
	}
	payload.CreatedBy = *utils.HexToMongoId(ctx, ctx.GetString("UserId"))
	payload.Email = ctx.GetString("Email")
	payload.Approved = false
	payload.Active = false

	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	for _, data := range payload.RequiredData {
		var invalid bool = true
		for _, d := range accessibleUserData {
			if d == data {
				invalid = false
			}
		}
		if invalid {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: fmt.Errorf("data field %s is not a valid user field to access", data), StatusCode: http.StatusBadRequest})
			return err
		}
	}
	_, err = appRepoInstance.CreateOne(payload)
	if err != nil {
		return err
	}
	return nil
}

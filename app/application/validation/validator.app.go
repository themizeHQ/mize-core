package application

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	appModel "mize.app/app/application/models"
	userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
)

func ValidateID(ctx *gin.Context, app *appModel.Application) error {
	var err error
	createdBy := userRepo.UserRepository.FindOneByFilter(ctx, map[string]interface{}{"id": app.CreatedBy})
	if createdBy == nil {
		err = errors.New("account assigned to the createdBy id does ot exist")
		app_errors.ErrorHandler(ctx, err, http.StatusNotFound)
	}
	return err
}

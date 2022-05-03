package application

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson/primitive"
	userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
)

type Application struct {
	Email             string
	CreatedBy         primitive.ObjectID
	Name              string
	Description       string
	LanguageAvailable []string
	Region            string
	Version           float32
	WorkSpaceOnly     primitive.ObjectID
	RegionAvailable   []string
	RequiredData      []string
	// ---  not provided by user ---
	Rating    int
	UserCount int
	CreatedAt primitive.Timestamp `bson:"createdAt"`
	UpdatedAt primitive.Timestamp `bson:"updatedAt"`
}

func (app *Application) MarshalBinary() ([]byte, error) {
	return json.Marshal(app)
}

func (app *Application) Validate() error {
	return validation.ValidateStruct(app, validation.Field(app.Email, validation.Required.Error("Please provide a valid email for your app"), is.Email.Error("Pass in a valid email")),
		validation.Field(app.CreatedBy, validation.Required, is.MongoID.Error("Pass in a valid mongodb id")),
		validation.Field(app.Name, validation.Required.Error("Please provide a name for your app"), is.Alphanumeric.Error("Your app name can contain only numbers and letters")),
		validation.Field(app.LanguageAvailable, validation.Each(validation.Required.Error("Provide at lease one langauge"))),
		validation.Field(app.Region, is.CountryCode2.Error("Pass in a valid country code")),
		validation.Field(app.WorkSpaceOnly, is.MongoID.Error("Pass in a WorkSpace ID")),
		validation.Field(app.RegionAvailable, is.CountryCode2.Error("Pass in a valid country code")),
		validation.Field(app.RequiredData, validation.Required.Error("Pass in the user information you need")),
	)
}

func (app *Application) ValidateID(ctx *gin.Context) error {
	var err error
	createdBy := userRepo.UserRepository.FindOneByFilter(ctx, map[string]interface{}{"id": app.CreatedBy})
	if createdBy == nil {
		err = errors.New("account assigned to the createdBy id does ot exist")
		app_errors.ErrorHandler(ctx, err, http.StatusNotFound)
	}
	return err
}

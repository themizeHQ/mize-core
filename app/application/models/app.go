package models

import (
	"encoding/json"
	"time"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Application struct {
	Id                primitive.ObjectID `bson:"_id"`
	Email             string             `bson:"email"`
	CreatedBy         primitive.ObjectID `bson:"createdBy"`
	Name              string             `bson:"name"`
	Description       string             `bson:"description"`
	LanguageAvailable []string           `bson:"languageAvailable"`
	Region            string             `bson:"region"`
	Version           string             `bson:"version"`
	WorkSpaceOnly     *string            `bson:"workspaceOnly"`
	RegionAvailable   []string           `bson:"regionAvailable"`
	RequiredData      []string           `bson:"requiredData"`
	Approved          bool               `bson:"approved"`
	Active            bool               `bson:"active"`
	// ---  not provided by user ---
	Rating    int
	UserCount int
	CreatedAt primitive.DateTime `bson:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt"`
}

func (app *Application) MarshalBinary() ([]byte, error) {
	if app.CreatedAt.Time().IsZero() {
		app.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	app.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return json.Marshal(app)
}

func (app *Application) MarshalBSON() ([]byte, error) {
	if app.CreatedAt.Time().Unix() == 0 {
		app.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	app.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*app)
}

func (channel Application) MongoDBName() string {
	return "Apps"
}

func (app *Application) Validate() error {
	return validation.ValidateStruct(app,
		validation.Field(&app.Email, validation.Required.Error("Please provide a valid email for your app"), is.Email.Error("Pass in a valid email")),
		validation.Field(&app.Name, validation.Required.Error("Please provide a name for your app"), is.Alphanumeric.Error("Your app name can contain only numbers and letters")),
		validation.Field(&app.LanguageAvailable, validation.Each(validation.Required.Error("Provide at lease one langauge"))),
		validation.Field(&app.Region, is.CountryCode2.Error("Pass in a valid country code")),
		validation.Field(&app.RegionAvailable, is.CountryCode2.Error("Pass in a valid country code")),
		validation.Field(&app.RequiredData, validation.Required.Error("Pass in the user information you need")),
	)
}

package models

import (
	"encoding/json"
	"time"

	"github.com/go-ozzo/ozzo-validation"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	Time int64  `bson:"time" json:"time"`
	Url  string `bson:"url" json:"url"`
}

type Schedule struct {
	Id          primitive.ObjectID  `bson:"_id" json:"id"`
	Name        string              `bson:"name" json:"name"`
	Location    string              `bson:"location" json:"location"`
	Details     string              `bson:"details" json:"details"`
	Importance  string              `bson:"importance" json:"importance"`
	Events      []Event             `bson:"time" json:"time"`
	WorkspaceId *primitive.ObjectID `bson:"workspaceId" json:"workspaceId"`
	CreatedBy   primitive.ObjectID  `bson:"createdBy" json:"createdBy"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

func (sch Schedule) MarshalBinary() ([]byte, error) {
	return json.Marshal(sch)
}

func (sch *Schedule) MarshalBSON() ([]byte, error) {
	if sch.CreatedAt.Time().Unix() == 0 {
		sch.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	sch.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*sch)
}

func (sch *Schedule) Validate() error {
	return validation.ValidateStruct(sch,
		validation.Field(&sch.Name, validation.Required.Error("name is a required field")),
		validation.Field(&sch.Location, validation.Required.Error("location name is a required field")),
		validation.Field(&sch.Details, validation.Required.Error("details is a required field")),
		validation.Field(&sch.Importance, validation.Required.Error("importance is a required field")),
		validation.Field(&sch.Events, validation.Required.Error("pass in at least 1 event")),
	)
}

func (channel Schedule) MongoDBName() string {
	return "Schedules"
}

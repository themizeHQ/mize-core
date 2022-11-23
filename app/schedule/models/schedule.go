package models

import (
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	notification_constants "mize.app/constants/notification"
)

type Event struct {
	Time int64  `bson:"time" json:"time"`
	Url  string `bson:"url" json:"url"`
}

type RecipientType string

var UserRecipient RecipientType = "user_recipient"
var TeamRecipient RecipientType = "team_recipient"

type Recipients struct {
	Name        string             `bson:"name" json:"name"`
	RecipientId primitive.ObjectID `bson:"recipientId" json:"recipientId"`
	Type        RecipientType      `bson:"type" json:"type"`
}

type Schedule struct {
	Id            primitive.ObjectID                                 `bson:"_id" json:"id"`
	Name          string                                             `bson:"name" json:"name"`
	Time          int64                                              `bson:"time" json:"time"`
	Location      string                                             `bson:"location" json:"location"`
	Importance    notification_constants.NotificationImportanceLevel `bson:"importance" json:"importance"`
	Url           *string                                            `bson:"url" json:"url"`
	Details       string                                             `bson:"details" json:"details"`
	From          string                                             `bson:"from" json:"from"`
	WorkspaceId   *primitive.ObjectID                                `bson:"workspaceId" json:"workspaceId"`
	CreatedBy     primitive.ObjectID                                 `bson:"createdBy" json:"createdBy"`
	Recipients    *[]Recipients                                      `bson:"recipient" json:"recipient"`
	RemindByEmail bool                                               `bson:"remindByEmail" json:"remindbyEmail"`
	RemindBySMS   bool                                               `bson:"remindBySMS" json:"remindbySMS"`

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
		validation.Field(&sch.Recipients, validation.Required.Error("pass in at least 1 recipient")),
	)
}

func (channel Schedule) MongoDBName() string {
	return "Schedules"
}

package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	notification_constants "mize.app/constants/notification"
)

type Alert struct {
	Id          primitive.ObjectID                                 `bson:"_id"`
	WorkspaceId primitive.ObjectID                                 `bson:"workspaceId"`
	UserIds     []primitive.ObjectID                               `bson:"usersId"`
	AdminId     primitive.ObjectID                                 `bson:"adminId"`
	ResourceId  *primitive.ObjectID                                `bson:"resourceId"`
	Importance  notification_constants.NotificationImportanceLevel `bson:"importance"`
	Message     string                                             `bson:"message"`

	CreatedAt primitive.DateTime `bson:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt"`
}

func (alert *Alert) MarshalBinary() ([]byte, error) {
	return json.Marshal(alert)
}

func (alert *Alert) MarshalBSON() ([]byte, error) {
	if alert.CreatedAt.Time().Unix() == 0 {
		alert.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	alert.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*alert)
}

func (alert Alert) MongoDBName() string {
	return "Alerts"
}

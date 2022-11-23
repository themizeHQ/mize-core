package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	notification_constants "mize.app/constants/notification"
)

type Alert struct {
	Id          primitive.ObjectID                                 `bson:"_id" json:"id"`
	WorkspaceId primitive.ObjectID                                 `bson:"workspaceId" json:"workspaceId"`
	UserIds     []primitive.ObjectID                               `bson:"usersId" json:"userId"`
	AdminId     primitive.ObjectID                                 `bson:"adminId" json:"adminId"`
	ResourceId  *primitive.ObjectID                                `bson:"resourceId" json:"resourceId"`
	ResourceUrl *string                                             `bson:"resourceUrl" json:"resourceUrl"`
	Importance  notification_constants.NotificationImportanceLevel `bson:"importance" json:"importance"`
	Message     string                                             `bson:"message" json:"message"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
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

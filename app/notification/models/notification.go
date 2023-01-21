package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	notification_constants "mize.app/constants/notification"
)

type Notification struct {
	Id          primitive.ObjectID                                 `bson:"_id" json:"id"`
	WorkspaceId *primitive.ObjectID                                `bson:"workspaceId" json:"workspaceId"`
	UserId      *primitive.ObjectID                                `bson:"userId" json:"userId"`
	ResourceId  *primitive.ObjectID                                `bson:"resourceId" json:"resourceId"`
	ExternalUrl *string                                            `bson:"externalURL" json:"externalURL"`
	Scope       notification_constants.NotificationScope           `bson:"scope" json:"scope"`
	Importance  notification_constants.NotificationImportanceLevel `bson:"importance" json:"importance"`
	Type        notification_constants.NotificationType            `bson:"type" json:"type"`
	Header      string                                             `bson:"header" json:"header"`
	Message     string                                             `bson:"message" json:"message"`
	Reacted     *bool                                              `bson:"reacted" json:"reacted"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

func (notification *Notification) MarshalBinary() ([]byte, error) {
	return json.Marshal(notification)
}

func (notification *Notification) MarshalBSON() ([]byte, error) {
	if notification.CreatedAt.Time().Unix() == 0 {
		notification.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	notification.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*notification)
}

func (notification Notification) MongoDBName() string {
	return "Notifications"
}

package notification

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	constnotif "mize.app/constants/notification"
)

type Notification struct {
	Id          primitive.ObjectID                     `bson:"_id"`
	WorkspaceId *primitive.ObjectID                     `bson:"workspaceId"`
	UserId      *primitive.ObjectID                     `bson:"userId"`
	ResourceId  primitive.ObjectID                     `bson:"resourceId"`
	Scope       constnotif.NotificationScope           `bson:"scope"`
	Importance  constnotif.NotificationImportanceLevel `bson:"importance"`
	Type        constnotif.NotificationType            `bson:"type"`
	Message     string                                 `bson:"message"`

	CreatedAt primitive.DateTime `bson:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt"`
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

package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Channel struct {
	Id           primitive.ObjectID `bson:"_id"`
	WorkspaceId  primitive.ObjectID `bson:"workspaceId"`
	Name         string             `bson:"name"`
	Description  string             `bson:"description"`
	Private      bool               `bson:"private"`
	CreatedBy    primitive.ObjectID `bson:"createdBy"`
	Compulsory   bool               `bson:"compulsory"`
	ProfileImage *string            `bson:"profileImage"`

	CreatedAt primitive.DateTime `bson:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt"`
}

func (channel *Channel) MarshalBinary() ([]byte, error) {
	return json.Marshal(channel)
}

func (channel *Channel) MarshalBSON() ([]byte, error) {
	if channel.CreatedAt.Time().Unix() == 0 {
		channel.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	channel.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*channel)
}

func (channel Channel) MongoDBName() string {
	return "Channels"
}

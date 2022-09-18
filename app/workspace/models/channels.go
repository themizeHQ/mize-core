package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Channel struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	WorkspaceId  primitive.ObjectID `bson:"workspaceId" json:"workspaceId"`
	Name         string             `bson:"name" json:"name"`
	Description  string             `bson:"description" json:"description"`
	Private      bool               `bson:"private" json:"private"`
	CreatedBy    primitive.ObjectID `bson:"createdBy" json:"createdBy"`
	Compulsory   bool               `bson:"compulsory" json:"compulsory"`
	ProfileImage *string            `bson:"profileImage" json:"profileImage"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
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

package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	Id             primitive.ObjectID  `bson:"_id"`
	To             primitive.ObjectID  `bson:"to"`
	From           primitive.ObjectID  `bson:"from"`
	WorkspaceId    primitive.ObjectID  `bson:"workspaceId"`
	Text           string              `bson:"text"`
	ReactionsCount int                 `bson:"reactionsCount"`
	ReplyTo        *primitive.ObjectID `bson:"replyTo"`
	ReplyCount     int                 `bson:"replyCount"`
	Username       string              `bson:"userName"`

	CreatedAt primitive.DateTime `bson:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt"`
}

func (message *Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(message)
}

func (message *Message) MarshalBSON() ([]byte, error) {
	if message.CreatedAt.Time().Unix() == 0 {
		message.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	message.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*message)
}

func (message Message) MongoDBName() string {
	return "Messages"
}

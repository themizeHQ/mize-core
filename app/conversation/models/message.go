package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mize.app/constants/message"
)

type Message struct {
	Id             primitive.ObjectID  `bson:"_id" json:"id"`
	To             primitive.ObjectID  `bson:"to" json:"to"`
	From           primitive.ObjectID  `bson:"from" json:"from"`
	WorkspaceId    *primitive.ObjectID  `bson:"workspaceId" json:"workspaceId"`
	Text           string              `bson:"text" json:"text"`
	ReactionsCount int                 `bson:"reactionsCount" json:"reactionCount"`
	ReplyTo        *primitive.ObjectID `bson:"replyTo" json:"replyTo"`
	ReplyCount     int                 `bson:"replyCount" json:"replyCount"`
	Username       string              `bson:"userName" json:"userName"`
	Type           message.MessageType `bson:"type" json:"type"`
	ResourceUrl    *string             `bson:"resourceUrl" json:"resourceUrl"`
	ProfileImage   string              `bson:"profileImage" json:"profileImage"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
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

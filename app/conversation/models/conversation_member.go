package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationMember struct {
	Id             primitive.ObjectID `bson:"_id" json:"id"`
	User           primitive.ObjectID `bson:"user" json:"user"`
	UserId         primitive.ObjectID `bson:"userId" json:"userId"`
	ReciepientName string             `bson:"reciepientName" json:"reciepientName"`
	ReciepientId   string             `bson:"reciepientId" json:"reciepientId"`
	Pinned         bool               `bson:"pinned" json:"pinned"`
	LastMessage    string             `bson:"lastMessage" json:"lastMessage"`
	LastSent       primitive.DateTime `bson:"lastMessageSent" json:"lastMessageSent"`
	UnreadMessages int                `bson:"unreadMessages" json:"unreadMessage"`
	ProfileImage   *string            `bson:"profileImage" json:"profileImage"`
	CreatedAt      primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt      primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

func (conv *ConversationMember) MarshalBinary() ([]byte, error) {
	return json.Marshal(conv)
}

func (conv *ConversationMember) MarshalBSON() ([]byte, error) {
	if conv.CreatedAt.Time().Unix() == 0 {
		conv.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	conv.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*conv)
}

func (conv ConversationMember) MongoDBName() string {
	return "ConversationMembers"
}

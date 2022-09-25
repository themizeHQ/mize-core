package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationMember struct {
	Id             primitive.ObjectID  `bson:"_id" json:"id"`
	ConversationId primitive.ObjectID  `bson:"conversationId" json:"conversationId"`
	UserId         primitive.ObjectID  `bson:"userId" json:"userId"`
	ReciepientName string              `bson:"reciepientName" json:"reciepientName"`
	ReciepientId   primitive.ObjectID  `bson:"reciepientId" json:"reciepientId"`
	Pinned         bool                `bson:"pinned" json:"pinned"`
	LastMessage    string              `bson:"lastMessage" json:"lastMessage"`
	LastSent       primitive.DateTime  `bson:"lastMessageSent" json:"lastMessageSent"`
	WorkspaceConv  bool                `bson:"workspaceConv" json:"workspaceConv"`
	WorkspaceId    *primitive.ObjectID `bson:"workspaceId" json:"workspaceId"`
	UnreadMessages int                 `bson:"unreadMessages" json:"unreadMessage"`
	ProfileImage   string              `bson:"profileImage" json:"profileImage"`
	CreatedAt      primitive.DateTime  `bson:"createdAt" json:"createdAt"`
	UpdatedAt      primitive.DateTime  `bson:"updatedAt" json:"updatedAt"`
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

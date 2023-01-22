package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Reaction struct {
	Id             primitive.ObjectID `bson:"_id" json:"id"`
	MessageID      primitive.ObjectID `bson:"messageId" json:"messageId"`
	UserID         primitive.ObjectID `bson:"userId" json:"userId"`
	ConversationID primitive.ObjectID `bson:"conversationId" json:"conversationId"`
	WorkspaceId    primitive.ObjectID `bson:"workspaceId" json:"workspaceId"`
	UserName       string             `bson:"userName" json:"userName"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

func (reaction *Reaction) MarshalBinary() ([]byte, error) {
	return json.Marshal(reaction)
}

func (reaction *Reaction) MarshalBSON() ([]byte, error) {
	if reaction.CreatedAt.Time().Unix() == 0 {
		reaction.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	reaction.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*reaction)
}

func (reaction Reaction) MongoDBName() string {
	return "reactions"
}

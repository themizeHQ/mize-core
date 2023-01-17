package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Conversation struct {
	Id            primitive.ObjectID   `bson:"_id" json:"id"`
	Participants  []primitive.ObjectID `bson:"participants" json:"participants"`
	WorkspaceConv bool                 `bson:"workspaceConv" json:"workspaceConv"`
	Pinned        bool                 `bson:"pinned" json:"pinned"`
	WorkspaceId   *primitive.ObjectID  `bson:"workspaceId" json:"workspaceId"`
	CreatedAt     primitive.DateTime   `bson:"createdAt" json:"createdAt"`
	UpdatedAt     primitive.DateTime   `bson:"updatedAt" json:"updatedAt"`
}

func (conv *Conversation) MarshalBinary() ([]byte, error) {
	return json.Marshal(conv)
}

func (conv *Conversation) MarshalBSON() ([]byte, error) {
	if conv.CreatedAt.Time().Unix() == 0 {
		conv.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	conv.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*conv)
}

func (conv Conversation) MongoDBName() string {
	return "Conversations"
}

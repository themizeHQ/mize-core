package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type DeleteMessage struct {
	MessageId primitive.ObjectID `bson:"_id"`
	To        primitive.ObjectID `bson:"to"`
}

type Reaction struct {
	MessageID      primitive.ObjectID `json:"messageID"`
	Reaction       string             `json:"reaction"`
	ConversationID primitive.ObjectID `json:"conversationId"`
}

type RemoveReaction struct {
	MessageID      primitive.ObjectID `json:"messageID"`
	ConversationID primitive.ObjectID `json:"conversationId"`
}

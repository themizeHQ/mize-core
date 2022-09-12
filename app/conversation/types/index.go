package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type DeleteMessage struct {
	MessageId primitive.ObjectID `bson:"_id"`
	To        primitive.ObjectID `bson:"to"`
}

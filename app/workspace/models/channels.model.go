package workspace

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Channel struct {
	Id          primitive.ObjectID   `bson:"_id"`
	WorkspaceId primitive.ObjectID   `bson:"workspaceId"`
	Name        string               `bson:"name"`
	Description string               `bson:"description"`
	Private     bool                 `bson:"private"`
	CreatedBy   primitive.ObjectID   `bson:"createdBy"`
	Admins      []primitive.ObjectID `bson:"admins"`
	Compulsory  bool                 `bson:"compulsory"`
}

func (channel *Channel) MarshalBinary() ([]byte, error) {
	return json.Marshal(channel)
}

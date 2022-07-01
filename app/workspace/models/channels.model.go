package workspace

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Channel struct {
	Id          primitive.ObjectID `bson:"_id"`
	WorkspaceId string             `bson:"workspaceId"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Private     bool               `bson:"private"`
	CreatedBy   string             `bson:"createdBy"`
	Admins      string             `bson:"admins"`
	Compulsory  bool               `bson:"compulsory"`

	CreatedAt primitive.Timestamp `bson:"createdAt"`
	UpdatedAt primitive.Timestamp `bson:"updatedAt"`
}

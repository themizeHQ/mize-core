package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TeamMembers struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	FirstName    string             `bson:"firstName" json:"firstName"`
	LastName     string             `bson:"lastName" json:"lastName"`
	UserId       primitive.ObjectID `bson:"userId" json:"userId"`
	WorkspaceId  primitive.ObjectID `bson:"workspaceId" json:"workspaceId"`
	TeamId       primitive.ObjectID `bson:"teamId" json:"teamId"`
	TeamName     string             `bson:"teamName" json:"teamName"`
	ProfileImage *string            `bson:"profileImage" json:"profileImage"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

func (members TeamMembers) MarshalBinary() ([]byte, error) {
	return json.Marshal(members)
}

func (members *TeamMembers) MarshalBSON() ([]byte, error) {
	if members.CreatedAt.Time().Unix() == 0 {
		members.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	members.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*members)
}

func (members TeamMembers) MongoDBName() string {
	return "TeamMembers"
}

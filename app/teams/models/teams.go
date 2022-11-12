package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Team struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Description  string             `bson:"description" json:"description"`
	WorkspaceId  primitive.ObjectID `bson:"workspaceId" json:"workspaceId"`
	MembersCount int                `bson:"membersCount" json:"membersCount"`
	ProfileImage *string            `bson:"profileImage" json:"profileImage"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

func (team Team) MarshalBinary() ([]byte, error) {
	return json.Marshal(team)
}

func (team *Team) MarshalBSON() ([]byte, error) {
	if team.CreatedAt.Time().Unix() == 0 {
		team.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	team.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*team)
}

func (team Team) MongoDBName() string {
	return "Teams"
}

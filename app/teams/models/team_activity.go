package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mize.app/app/teams/types"
)

type TeamActivity struct {
	Id           primitive.ObjectID     `bson:"_id" json:"id"`
	WorkspaceId  primitive.ObjectID     `bson:"workspaceId" json:"workspaceId"`
	TeamId       primitive.ObjectID     `bson:"teamId" json:"teamId"`
	ResourceID   primitive.ObjectID     `bson:"resourceID" json:"resourceID"`
	TeamName     string                 `bson:"teamName" json:"teamName"`
	ProfileImage *string                `bson:"profileImage" json:"profileImage"`
	Name         string                 `bson:"name" json:"name"`
	Type         types.TeamActivityType `bson:"type" json:"type"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

func (acticity TeamActivity) MarshalBinary() ([]byte, error) {
	return json.Marshal(acticity)
}

func (acticity *TeamActivity) MarshalBSON() ([]byte, error) {
	if acticity.CreatedAt.Time().Unix() == 0 {
		acticity.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	acticity.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*acticity)
}

func (acticity TeamActivity) MongoDBName() string {
	return "TeamActivity"
}

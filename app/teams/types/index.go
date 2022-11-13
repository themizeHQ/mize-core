package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type TeamMemberType string

var TeamType TeamMemberType = "team_type"
var UserType TeamMemberType = "user_type"

type TeamMembersPayload struct {
	UserID primitive.ObjectID `bson:"userId" json:"userId"`
	Type   TeamMemberType     `bson:"type" json:"type"`
}

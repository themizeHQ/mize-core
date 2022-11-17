package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type TeamMemberType string

var TeamType TeamMemberType = "team_type"
var UserType TeamMemberType = "user_type"

type TeamMembersPayload struct {
	UserID primitive.ObjectID `bson:"userId" json:"userId"`
	Type   TeamMemberType     `bson:"type" json:"type"`
}

type IDArray []primitive.ObjectID

type TeamActivityType string

var AddToChannel TeamActivityType = "add_to_channel"
var AddToTeam TeamActivityType = "add_to_team"
var AddToSchedule TeamActivityType = "add_to_schedule"

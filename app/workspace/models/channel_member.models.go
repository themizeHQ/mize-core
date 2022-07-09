package workspace

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChannelAdminAccess string

const (
	CHANNEL_FULL_ACCESS AdminAccess = "full_access"
)

type ChannelMemberActions string

type ChannelMember struct {
	Id          primitive.ObjectID     `bson:"_id"`
	ChannelId   string                 `bson:"workspaceId"`
	Username    string                 `bson:"userName"`
	UserId      string                 `bson:"userId"`
	Admin       bool                   `bson:"admin"`
	AdminAccess []ChannelAdminAccess   `bson:"adminAccess"`
	JoinDate    int64                  `bson:"joinDate"`
	Banned      bool                   `bson:"banned"`
	Restricted  []ChannelMemberActions `bson:"restricted"`
}

func (member *ChannelMember) MarshalBinary() ([]byte, error) {
	return json.Marshal(member)
}

func (member *ChannelMember) Validate() error {
	return validation.ValidateStruct(member,
		validation.Field(&member.Username, validation.Required.Error("provide username of member")),
		validation.Field(&member.ChannelId, validation.Required.Error("provide workspace of member"), is.MongoID),
		validation.Field(&member.UserId, validation.Required.Error("prodvide a userid")),
		validation.Field(&member.JoinDate, validation.Required.Error("join date is required")),
	)
}





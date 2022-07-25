package workspace

import (
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChannelAdminAccess string

const (
	CHANNEL_FULL_ACCESS ChannelAdminAccess = "full_access"
)

type ChannelMemberActions string

type ChannelMember struct {
	Id          primitive.ObjectID     `bson:"_id"`
	ChannelId   primitive.ObjectID     `bson:"channelId"`
	WorkspaceId primitive.ObjectID     `bson:"workspaceId"`
	Username    string                 `bson:"userName"`
	UserId      primitive.ObjectID     `bson:"userId"`
	Admin       bool                   `bson:"admin"`
	AdminAccess []ChannelAdminAccess   `bson:"adminAccess"`
	JoinDate    primitive.DateTime     `bson:"joinDate"`
	Banned      bool                   `bson:"banned"`
	Restricted  []ChannelMemberActions `bson:"restricted"`
	CreatedAt   primitive.DateTime     `bson:"createdAt"`
	UpdatedAt   primitive.DateTime     `bson:"updatedAt"`
}

func (member *ChannelMember) MarshalBinary() ([]byte, error) {
	return json.Marshal(member)
}

func (member *ChannelMember) MarshalBSON() ([]byte, error) {
	if member.CreatedAt.Time().Unix() == 0 {
		member.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	member.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*member)
}

func (member *ChannelMember) Validate() error {
	return validation.ValidateStruct(member,
		validation.Field(&member.Username, validation.Required.Error("provide username of member")),
		validation.Field(&member.ChannelId, validation.Required.Error("provide workspace of member")),
		validation.Field(&member.UserId, validation.Required.Error("prodvide a userid")),
		validation.Field(&member.JoinDate, validation.Required.Error("join date is required")),
	)
}

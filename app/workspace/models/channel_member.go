package models

import (
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"mize.app/constants/channel"
)

type ChannelMemberActions string

type ChannelMember struct {
	Id             primitive.ObjectID           `bson:"_id" json:"id"`
	ChannelId      primitive.ObjectID           `bson:"channelId" json:"channelId"`
	ChannelName    string                       `bson:"channelName" json:"channelName"`
	WorkspaceId    primitive.ObjectID           `bson:"workspaceId" json:"workspaceId"`
	Username       string                       `bson:"userName" json:"userName"`
	UserId         primitive.ObjectID           `bson:"userId" json:"userId"`
	Admin          bool                         `bson:"admin" json:"admin"`
	AdminAccess    []channel.ChannelAdminAccess `bson:"adminAccess" json:"adminAccess"`
	JoinDate       primitive.DateTime           `bson:"joinDate" json:"joinDate"`
	Banned         bool                         `bson:"banned" json:"banned"`
	Restricted     []ChannelMemberActions       `bson:"restricted" json:"restricted"`
	Pinned         bool                         `bson:"pinned" json:"pinned"`
	LastMessage    string                       `bson:"lastMessage" json:"lastMessage"`
	LastSent       primitive.DateTime           `bson:"lastMessageSent" json:"lastMessageSent"`
	UnreadMessages int                          `bson:"unreadMessages" json:"unreadMessage"`
	ProfileImage   *string                      `bson:"profileImage" json:"profileImage"`
	CreatedAt      primitive.DateTime           `bson:"createdAt" json:"createdAt"`
	UpdatedAt      primitive.DateTime           `bson:"updatedAt" json:"updatedAt"`
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

func (channel ChannelMember) MongoDBName() string {
	return "ChannelMembers"
}

func (member *ChannelMember) Validate() error {
	return validation.ValidateStruct(member,
		validation.Field(&member.Username, validation.Required.Error("provide username of member")),
		validation.Field(&member.ChannelId, validation.Required.Error("provide workspace of member")),
		validation.Field(&member.UserId, validation.Required.Error("prodvide a userid")),
		validation.Field(&member.JoinDate, validation.Required.Error("join date is required")),
	)
}

func (member *ChannelMember) HasAccess(access_to_check []channel.ChannelAdminAccess) bool {
	has_access := false
	for _, user_access := range member.AdminAccess {
		for _, access := range access_to_check {
			if access == user_access {
				has_access = true
				break
			}
		}
	}
	return has_access
}

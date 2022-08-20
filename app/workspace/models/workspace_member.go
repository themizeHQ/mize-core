package models

import (
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	workspace_member_constants "mize.app/constants/workspace"
)

type WorkspaceMember struct {
	Id          primitive.ObjectID                            `bson:"_id"`
	WorkspaceId primitive.ObjectID                            `bson:"workspaceId"`
	Username    string                                        `bson:"userName"`
	UserId      primitive.ObjectID                            `bson:"userId"`
	Admin       bool                                          `bson:"admin"`
	AdminAccess []workspace_member_constants.AdminAccessType  `bson:"adminAccess"`
	JoinDate    int64                                         `bson:"joinDate"`
	Banned      bool                                          `bson:"banned"`
	Restricted  []workspace_member_constants.MemberActionType `bson:"restricted"`

	CreatedAt primitive.DateTime `bson:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt"`
}

func (member *WorkspaceMember) MarshalBinary() ([]byte, error) {
	return json.Marshal(member)
}

func (member *WorkspaceMember) MarshalBSON() ([]byte, error) {
	if member.CreatedAt.Time().Unix() == 0 {
		member.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	member.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*member)
}

func (channel WorkspaceMember) MongoDBName() string {
	return "WorkspaceMembers"
}

func (member *WorkspaceMember) Validate() error {
	return validation.ValidateStruct(member,
		validation.Field(&member.Username, validation.Required.Error("provide username of member")),
		validation.Field(&member.JoinDate, validation.Required.Error("join date is required")),
	)
}

func (member *WorkspaceMember) HasAccess(access_to_check []workspace_member_constants.AdminAccessType) bool {
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

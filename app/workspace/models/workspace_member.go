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
	Id                    primitive.ObjectID                            `bson:"_id" json:"id"`
	WorkspaceId           primitive.ObjectID                            `bson:"workspaceId" json:"workspace"`
	WorkspaceName         string                                        `bson:"workspaceName" json:"workspaceName"`
	Username              string                                        `bson:"userName" json:"userName"`
	Firstname             string                                        `bson:"firstName" json:"firstName"`
	Lastname              string                                        `bson:"lastName" json:"lastName"`
	UserId                primitive.ObjectID                            `bson:"userId" json:"userId"`
	Admin                 bool                                          `bson:"admin" json:"admin"`
	AdminAccess           []workspace_member_constants.AdminAccessType  `bson:"adminAccess" json:"adminAccess"`
	JoinDate              int64                                         `bson:"joinDate" json:"joinDate"`
	Banned                bool                                          `bson:"banned" json:"banned"`
	Deactivated           bool                                          `bson:"deactivated" json:"deactivated"`
	Restricted            []workspace_member_constants.MemberActionType `bson:"restricted" json:"restricted"`
	ProfileImage          *string                                       `bson:"profileImage" json:"profileImage"`
	ProfileImageThumbNail *string                                       `bson:"profileImageThumbnail" json:"profileImageThumbnail"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
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

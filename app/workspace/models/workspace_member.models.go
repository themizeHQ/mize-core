package workspace

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminAccess string

const (
	FULL_ACCESS AdminAccess = "full_access"
)

type MemberActions string

type WorkspaceMember struct {
	Id          primitive.ObjectID `bson:"_id"`
	WorkspaceId string             `bson:"workspaceId"`
	Username    string             `bson:"userName"`
	UserId      string             `bson:"userId"`
	Admin       bool               `bson:"admin"`
	AdminAccess []AdminAccess      `bson:"adminAccess"`
	JoinDate    int64              `bson:"joinDate"`
	Banned      bool               `bson:"banned"`
	Restricted  []MemberActions    `bson:"restricted"`
}

func (member *WorkspaceMember) MarshalBinary() ([]byte, error) {
	return json.Marshal(member)
}

func (member *WorkspaceMember) Validate() error {
	return validation.ValidateStruct(member,
		validation.Field(&member.Username, validation.Required.Error("provide username of member")),
		validation.Field(&member.WorkspaceId, validation.Required.Error("provide workspace of member"), is.MongoID),
		validation.Field(&member.UserId, validation.Required.Error("prodvide a userid")),
		validation.Field(&member.JoinDate, validation.Required.Error("join date is required")),
	)
}

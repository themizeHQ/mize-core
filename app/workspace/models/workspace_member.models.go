package workspace

import (
	"encoding/json"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminAccess string

const (
	FULL_ACCESS AdminAccess = "full_access"
)

type MemberActions string

type WorkspaceMember struct {
	Id          primitive.ObjectID `bson:"_id"`
	WorkspaceId primitive.ObjectID `bson:"workspaceId"`
	Username    string             `bson:"userName"`
	UserId      primitive.ObjectID `bson:"userId"`
	Admin       bool               `bson:"admin"`
	AdminAccess []AdminAccess      `bson:"adminAccess"`
	JoinDate    int64              `bson:"joinDate"`
	Banned      bool               `bson:"banned"`
	Restricted  []MemberActions    `bson:"restricted"`

	CreatedAt primitive.DateTime `bson:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt"`
}

func (member *WorkspaceMember) MarshalBinary() ([]byte, error) {
	return json.Marshal(member)
}

func (member *WorkspaceMember) MarshalBSON() ([]byte, error) {
	fmt.Println(member.CreatedAt.Time().Unix())
	if member.CreatedAt.Time().Unix() == 0 {
		member.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	member.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*member)
}

func (member *WorkspaceMember) Validate() error {
	return validation.ValidateStruct(member,
		validation.Field(&member.Username, validation.Required.Error("provide username of member")),
		validation.Field(&member.JoinDate, validation.Required.Error("join date is required")),
	)
}

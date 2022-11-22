package models

import (
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkspaceInvite struct {
	Id            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email         string             `bson:"email" json:"email"`
	WorkspaceName string             `bson:"workspaceName" json:"workspaceName"`
	Accepted      *bool              `bson:"accepted" json:"accepted"`
	Success       bool               `bson:"success" json:"success"`
	WorkspaceId   primitive.ObjectID `bson:"workspaceId" json:"workspaceId"`
	Expired       bool               `bson:"expired" json:"expired"`

	CreatedAt primitive.DateTime `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt,omitempty" json:"updatedAt"`
}

func (invite *WorkspaceInvite) MarshalBinary() ([]byte, error) {
	return json.Marshal(invite)
}

func (invite *WorkspaceInvite) MarshalBSON() ([]byte, error) {
	if invite.CreatedAt.Time().Unix() == 0 {
		invite.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	invite.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*invite)
}

func (channel WorkspaceInvite) MongoDBName() string {
	return "WorkspaceInvites"
}

func (invite *WorkspaceInvite) Validate() error {
	return validation.ValidateStruct(invite,
		validation.Field(&invite.Email, validation.Required.Error("Email is a required field"), is.Email.Error("Field must be a valid email")),
	)
}

package models

import (
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mize.app/constants/workspace"
)

type Workspace struct {
	Id           primitive.ObjectID      `bson:"_id" json:"id"`
	Name         string                  `bson:"name" json:"name"`
	Email        string                  `bson:"email" json:"email"`
	Description  string                  `bson:"description" json:"description"`
	Censor       bool                    `bson:"censor" json:"censor"`
	Type         workspace.WorkspaceType `bson:"type" json:"type"`
	CreatedBy    primitive.ObjectID      `bson:"createdBy" json:"createdBy"`
	ProfileImage *string                 `bson:"profileImage" json:"profileImage"`
	MemberCount  int                     `bson:"memberCount" json:"memberCount"`
	TeamsCount   int                     `bson:"teamCount" json:"teamCount"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

func (workspace *Workspace) MarshalBinary() ([]byte, error) {
	return json.Marshal(workspace)
}

func (workspace *Workspace) MarshalBSON() ([]byte, error) {
	if workspace.CreatedAt.Time().Unix() == 0 {
		workspace.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	workspace.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*workspace)
}

func (channel Workspace) MongoDBName() string {
	return "Workspaces"
}

func (workspace *Workspace) Validate() error {
	return validation.ValidateStruct(workspace,
		validation.Field(&workspace.Name, validation.Required.Error("Name is a required field")),
		validation.Field(&workspace.Email, validation.Required.Error("Email is a required field"), is.Email.Error("Field must be a valid email")),
	)
}

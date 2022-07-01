package workspace

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkspaceInvite struct {
	Id        primitive.ObjectID `bson:"_id"`
	Email     string
	Accepted  *bool
	Success   bool
	Workspace string
	Expired   bool

	CreatedAt primitive.Timestamp `bson:"createdAt"`
	UpdatedAt primitive.Timestamp `bson:"updatedAt"`
}

func (invite *WorkspaceInvite) MarshalBinary() ([]byte, error) {
	return json.Marshal(invite)
}

func (invite *WorkspaceInvite) Validate() error {
	return validation.ValidateStruct(invite,
		validation.Field(&invite.Email, validation.Required.Error("Email is a required field"), is.Email.Error("Field must be a valid email")),
	)
}

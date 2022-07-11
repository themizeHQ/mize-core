package workspace

import (
	"encoding/json"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Workspace struct {
	Id        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Censor    bool               `bson:"censor"`
	CreatedBy string             `bson:"createdBy"`

	CreatedAt primitive.Timestamp `bson:"createdAt"`
	UpdatedAt primitive.Timestamp `bson:"updatedAt"`
}

func (workspace *Workspace) MarshalBinary() ([]byte, error) {
	return json.Marshal(workspace)
}

func (workspace *Workspace) Validate() error {
	return validation.ValidateStruct(workspace,
		validation.Field(&workspace.Name, validation.Required.Error("Name is a required field")),
		validation.Field(&workspace.Email, validation.Required.Error("Email is a required field"), is.Email.Error("Field must be a valid email")),
	)
}

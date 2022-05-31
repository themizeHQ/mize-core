package user

import (
	"encoding/json"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"mize.app/cryptography"
)

type User struct {
	Id          string               `bson:"_id" json:"id"`
	FirstName   string               `bson:"firstName"`
	LastName    string               `bson:"lastName"`
	UserName    string               `bson:"userName"`
	Email       string               `bson:"email"`
	Region      string               `bson:"region"`
	Password    string               `bson:"password" json:"-"`
	Verified    bool                 `bson:"verified"`
	AppsCreated []primitive.ObjectID `bson:"appsCreated"`
	CreatedAt   primitive.Timestamp  `bson:"createdAt"`
	UpdatedAt   primitive.Timestamp  `bson:"updatedAt"`
}

func (user *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(user)
}

func (user *User) Validate() error {
	return validation.ValidateStruct(user,
		validation.Field(&user.FirstName, validation.Required.Error("FirstName is a required field")),
		validation.Field(&user.LastName, validation.Required.Error("LaststName is a required field")),
		validation.Field(&user.UserName, validation.Required.Error("UserName is a required field")),
		validation.Field(&user.Email, validation.Required.Error("Email is a required field"), is.Email.Error("Field must be a valid email")),
		validation.Field(&user.Region, validation.Required.Error("Please pass in your region"), is.CountryCode2.Error("Please pass in a valid country code")),
		validation.Field(&user.Password, validation.Length(6, 100).Error("Password cannot be less than 6 digits"), validation.Required.Error("Password is a required field")),
	)
}

func (user *User) RunHooks() {
	user.beforeInsertHook()
}

func (user *User) beforeInsertHook() {
	password := cryptography.HashString(user.Password, nil)
	user.Password = string(password)
}

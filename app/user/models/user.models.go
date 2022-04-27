package user

import (
	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	FirstName   string               `bson:"firstName"`
	LastName    string               `bson:"lastName"`
	UserName    string               `bson:"userName"`
	Email       string               `bson:"email"`
	Region      string               `bson:"region"`
	Password    string               `bson:"password"`
	Verified    bool                 `bson:"verified"`
	OrgsCreated []primitive.ObjectID `bson:"orgsCreated"`
	CreatedAt   primitive.Timestamp  `bson:"createdAt"`
	UpdatedAt   primitive.Timestamp  `bson:"updatedAt"`
}

func (user *User) Validate() error {
	return validation.ValidateStruct(user,
		validation.Field(&user.FirstName, validation.Required.Error("FirstName is a required field")),
		validation.Field(&user.LastName, validation.Required.Error("LaststName is a required field")),
		validation.Field(&user.UserName, validation.Required.Error("UserName is a required field")),
		validation.Field(&user.Email, validation.Required.Error("Email is a required field"), is.Email.Error("Field must be a valid email")),
		validation.Field(&user.Region, validation.Required.Error("Please pass in your region"), is.CountryCode2.Error("Please pass in a valid country code")),
		validation.Field(&user.Password, validation.Length(6, 30).Error("Password cannot be less than 6 digits"), validation.Required.Error("Password is a required field")),
	)
}

package user

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mize.app/cryptography"
)

type User struct {
	Id        primitive.ObjectID `bson:"_id"`
	FirstName string             `bson:"firstName"`
	LastName  string             `bson:"lastName"`
	UserName  string             `bson:"userName"`
	Email     string             `bson:"email"`
	Region    string             `bson:"region"`
	Password  string             `json:"-" bson:"password"`
	Verified  bool               `bson:"verified"`
	CreatedAt primitive.DateTime `bson:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt"`
}

func (user *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(user)
}

func (user *User) MarshalBSON() ([]byte, error) {
	fmt.Println(user.CreatedAt.Time().Unix())
	if user.CreatedAt.Time().Unix() == 0 {
		user.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	user.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*user)
}

func (user *User) Validate() error {
	return validation.ValidateStruct(user,
		validation.Field(&user.UserName, validation.Required.Error("UserName is a required field")),
		validation.Field(&user.Email, validation.Required.Error("Email is a required field"), is.Email.Error("Field must be a valid email")),
		validation.Field(&user.Region, is.CountryCode2.Error("Please pass in a valid country code")),
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

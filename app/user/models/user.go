package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	user_constants "mize.app/constants/user"
	"mize.app/cryptography"
)

type User struct {
	Id              primitive.ObjectID                   `bson:"_id"`
	FirstName       string                               `bson:"firstName"`
	LastName        string                               `bson:"lastName"`
	UserName        string                               `bson:"userName"`
	Email           string                               `bson:"email"`
	Region          string                               `bson:"region"`
	Password        string                               `bson:"password"`
	Verified        bool                                 `bson:"verified"`
	Language        string                               `bson:"language"`
	Phone           string                               `bson:"phone"`
	Status          user_constants.UserStatusType        `bson:"status"`
	ProfileImage    *string                              `bson:"profileImage"`
	ACSUserId       string                               `bson:"acsUserId"`
	Discoverability []user_constants.UserDiscoverability `bson:"discoverability"`

	CreatedAt primitive.DateTime `bson:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt"`
}

type UpdateUser struct {
	FirstName       string                               `json:"firstName" bson:"firstName,omitempty"`
	LastName        string                               `json:"lastName" bson:"lastName,omitempty"`
	Region          string                               `json:"region" bson:"region,omitempty"`
	Language        string                               `json:"language" bson:"language,omitempty"`
	Phone           *string                              `json:"phone" bson:"phone,omitempty"`
	Status          user_constants.UserStatusType        `json:"status" bson:"status,omitempty"`
	Discoverability []user_constants.UserDiscoverability `json:"discoverability" bson:"discoverability,omitempty"`
}

// func (user *User) MarshalBinary() ([]byte, error) {
// 	return json.Marshal(user)
// }

func (user User) MarshalBinary() ([]byte, error) {
	return json.Marshal(user)
}

func (user *User) MarshalBSON() ([]byte, error) {
	if user.CreatedAt.Time().Unix() == 0 {
		user.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	user.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*user)
}

func (user *User) Validate() error {
	fmt.Println(user.Language)
	return validation.ValidateStruct(user,
		validation.Field(&user.UserName, validation.Required.Error("username is a required field")),
		validation.Field(&user.LastName, validation.Required.Error("lastname is a required field")),
		validation.Field(&user.FirstName, validation.Required.Error("firstname is a required field")),
		validation.Field(&user.Email, validation.Required.Error("email is a required field"), is.Email.Error("proide a valid email")),
		validation.Field(&user.Region, is.CountryCode2.Error("provide a valid ISO3166 country code")),
		validation.Field(&user.Password, validation.Length(6, 100).Error("password cannot be less than 6 digits or more than 100"), validation.Required.Error("password is a required field")),
		validation.Field(&user.Language, validation.In(user_constants.AvailableUserLanguage...).Error("language selected is not available on mize")),
		validation.Field(&user.Status, validation.In(user_constants.AVAILABLE, user_constants.AWAY, user_constants.BUSY, user_constants.MEETING).Error("invalid status selected")),
	)
}

func (user *UpdateUser) ValidateUpdate() error {
	return validation.ValidateStruct(user,
		validation.Field(&user.Region, is.CountryCode2.Error("provide a valid ISO3166 country code")),
		validation.Field(&user.Phone, validation.NotIn("")),
		validation.Field(&user.Region, is.CountryCode2.Error("provide a valid ISO3166 country code")),
		validation.Field(&user.Language, validation.In(user_constants.AvailableUserLanguage...).Error("language selected is not available on mize")),
		validation.Field(&user.Discoverability, validation.Each(validation.In(user_constants.DISCOVERABILITY_EMAIL, user_constants.DISCOVERABILITY_PHONE, user_constants.DISCOVERABILITY_USERNAME).Error("invalid discoverability setting selected"))),
		validation.Field(&user.Status, validation.In(user_constants.AVAILABLE, user_constants.AWAY, user_constants.BUSY, user_constants.MEETING).Error("invalid status selected")),
	)
}

func (user *User) RunHooks() {
	user.beforeInsertHook()
}

func (user *User) beforeInsertHook() {
	password := cryptography.HashString(user.Password, nil)
	user.Password = string(password)
}

func (user *User) ValidatePassword() error {
	return validation.ValidateStruct(user,
		validation.Field(&user.Password, validation.Length(6, 100).Error("Password cannot be less than 6 digits"), validation.Required.Error("Password is a required field")),
	)
}

func (channel User) MongoDBName() string {
	return "Users"
}

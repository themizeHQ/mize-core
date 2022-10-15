package types

type PhoneNumberFilter struct {
	Phone []string `bson:"phone" json:"phone"`
}

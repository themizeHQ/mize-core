package utils

import "go.mongodb.org/mongo-driver/bson/primitive"

func StringToDbId(id string) primitive.ObjectID {
	parsed, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		panic("invalid id attempted to be parsed to objectid")
	}
	return parsed
}

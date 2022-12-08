package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CallLog struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	UserId       primitive.ObjectID `bson:"userId" json:"userId"`
	ProfileImage string             `bson:"profileImage" json:"profileImage"`
	Duration     int                `bson:"duration" json:"duration"`
	Time         primitive.DateTime `bson:"time" json:"time"`
	Dialed       bool               `bson:"dialed" json:"dialed"`
	Missed       bool               `bson:"missed" json:"missed"`
	CreatedAt    primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt    primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

func (log *CallLog) MarshalBinary() ([]byte, error) {
	return json.Marshal(log)
}

func (log *CallLog) MarshalBSON() ([]byte, error) {
	if log.CreatedAt.Time().Unix() == 0 {
		log.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	log.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*log)
}

func (log CallLog) MongoDBName() string {
	return "CallLogs"
}

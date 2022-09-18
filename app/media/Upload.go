package media

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Upload struct {
	Id       primitive.ObjectID `bson:"_id" json:"id"`
	Url      string             `bson:"url" json:"url"`
	Bytes    int                `bson:"bytes" json:"bytes"`
	FileName string             `bson:"fileName" json:"fileName"`
	Type     string             `bson:"type" json:"type"`
	PublicID string             `bson:"publicId" json:"publicId"`
	Service  string             `bson:"service" json:"service"`
	UploadBy primitive.ObjectID `bson:"uploadBy" json:"uploadBy"`
	Format   string             `bson:"format" json:"format"`

	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

func (upload *Upload) MarshalBinary() ([]byte, error) {
	return json.Marshal(upload)
}

func (upload *Upload) MarshalBSON() ([]byte, error) {
	if upload.CreatedAt.Time().Unix() == 0 {
		upload.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	upload.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return bson.Marshal(*upload)
}

func (upload Upload) MongoDBName() string {
	return "Upload"
}

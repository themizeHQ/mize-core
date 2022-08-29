package media

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Upload struct {
	Id       primitive.ObjectID `bson:"_id"`
	Url      string             `bson:"url"`
	Bytes    int                `bson:"bytes"`
	FileName string             `bson:"fileName"`
	Type     string             `bson:"type"`
	PublicID string             `bson:"publicId"`
	Service  string             `bson:"service"`
	UploadBy primitive.ObjectID `bson:"uploadBy"`
	Format   string             `bson:"format"`

	CreatedAt primitive.DateTime `bson:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt"`
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

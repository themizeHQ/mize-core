package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoModels interface {
	MongoDBName() string
}

type ModelMethods interface {
	MarshalBSON() ([]byte, error)
	MarshalBinary() ([]byte, error)
}

type MongoRepository[T MongoModels] struct {
	Model   *mongo.Collection
	Payload interface{}
}

func (repo *MongoRepository[T]) CreateOne(payload T, opts ...*options.InsertOneOptions) (*string, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()

	response, err := repo.Model.InsertOne(c, parsePayload(payload), opts...)
	if err != nil {
		return nil, err
	}
	id := response.InsertedID.(primitive.ObjectID).Hex()
	return &id, err
}

func (repo *MongoRepository[T]) CreateBulk(payload []T, opts ...*options.InsertManyOptions) (*[]string, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()
	parsed_payload := parseMultiple(payload)
	marshaled := []interface{}{}
	for _, i := range parsed_payload {
		interface{}(i).(ModelMethods).MarshalBSON()
		interface{}(i).(ModelMethods).MarshalBinary()
		marshaled = append(marshaled, i)
	}
	response, err := repo.Model.InsertMany(c, marshaled, opts...)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, id := range response.InsertedIDs {
		ids = append(ids, id.(primitive.ObjectID).Hex())
	}
	return &ids, err
}

func (repo *MongoRepository[T]) FindOneByFilter(filter map[string]interface{}, opts ...*options.FindOneOptions) (*T, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()
	var result T
	f := parseFilter(filter)
	doc := repo.Model.FindOne(c, f, opts...)
	err := doc.Decode(&result)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil, errors.New("no documents found")
		}
		return nil, err
	}
	return &result, nil
}

func (repo *MongoRepository[T]) FindMany(filter map[string]interface{}, opts ...*options.FindOptions) (*[]T, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()
	var result []T
	f := parseFilter(filter)
	cursor, err := repo.Model.Find(c, f, opts...)
	if err != nil {
		return nil, err
	}
	err = cursor.All(c, &result)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil, errors.New("no documents found")
		}
		return nil, err
	}
	return &result, nil
}

func (repo *MongoRepository[T]) FindById(id string, opts ...*options.FindOneOptions) (*T, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()
	var result T
	err := repo.Model.FindOne(c, bson.M{"_id": parseStringToMongo(&id)}, opts...).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (repo *MongoRepository[T]) CountDocs(filter map[string]interface{}, opts ...*options.CountOptions) (int64, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()
	cc :=  parseFilter(filter)
	fmt.Println(cc)
	count, err := repo.Model.CountDocuments(c, cc, opts...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *MongoRepository[T]) FindLast(opts ...*options.FindOptions) (*T, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()

	var lastRecord T
	err := repo.Model.FindOne(c, bson.M{}, options.FindOne().SetSort(bson.M{"$natural": -1})).Decode(&lastRecord)
	if err != nil {
		return nil, err
	}
	return &lastRecord, nil
}

func (repo *MongoRepository[T]) DeleteOne(filter map[string]interface{}) (bool, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()

	_, err := repo.Model.DeleteOne(c, parseFilter(filter))
	if err != nil {
		return false, err
	}
	return true, err
}

func (repo *MongoRepository[T]) DeleteById(id string) (bool, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()

	_, err := repo.Model.DeleteOne(c, bson.M{"_id": parseStringToMongo(&id)})
	if err != nil {
		return false, err
	}
	return true, err
}

func (repo *MongoRepository[T]) DeleteMany(ctx *gin.Context, filter map[string]interface{}) (bool, error) {
	_, err := repo.Model.DeleteMany(ctx, filter)
	if err != nil {
		return false, err
	}
	return true, err
}

func (repo *MongoRepository[T]) UpdateByField(filter map[string]interface{}, payload *T, opts ...*options.UpdateOptions) (bool, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()

	_, err := repo.Model.UpdateOne(c, parseFilter(filter), payload, opts...)
	if err != nil {
		return false, err
	}
	return true, err
}

func (repo *MongoRepository[T]) UpdateOrCreateByField(filter map[string]interface{}, payload map[string]interface{}, opts ...*options.UpdateOptions) (bool, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()

	_, err := repo.Model.UpdateOne(c, parseFilter(filter), bson.D{primitive.E{Key: "$set", Value: payload}}, opts...)
	if err != nil {
		return false, err
	}
	return true, err
}

func (repo *MongoRepository[T]) UpdateOrCreateByFieldAndReturn(filter map[string]interface{}, payload map[string]interface{}, opts ...*options.UpdateOptions) (*string, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()

	result, err := repo.Model.UpdateOne(c, parseFilter(filter), bson.D{primitive.E{Key: "$set", Value: payload}}, opts...)
	if err != nil {
		return nil, err
	}
	if result.UpsertedID == nil {
		return nil, nil
	}
	id := result.UpsertedID.(primitive.ObjectID).Hex()
	return &id, err
}

func (repo *MongoRepository[T]) UpdateById(ctx *gin.Context, id string, payload *T, opts ...*options.UpdateOptions) (bool, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()

	_, err := repo.Model.UpdateByID(c, parseStringToMongo(&id), bson.D{primitive.E{Key: "$set", Value: *payload}}, opts...)
	if err != nil {
		return false, err
	}
	return true, err
}

func (repo *MongoRepository[T]) UpdatePartialById(ctx *gin.Context, id string, payload interface{}, opts ...*options.UpdateOptions) (bool, error) {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()

	_, err := repo.Model.UpdateByID(c, parseStringToMongo(&id), bson.D{primitive.E{Key: "$set", Value: payload}}, opts...)
	if err != nil {
		return false, err
	}
	return true, err
}

func (repo *MongoRepository[T]) StartTransaction(ctx *gin.Context, payload func(mongo.SessionContext) error) error {
	c, cancel := createCtx()

	defer func() {
		cancel()
	}()

	if err := repo.Model.Database().Client().UseSession(c, func(sc mongo.SessionContext) error {
		if err := sc.StartTransaction(); err != nil {
			return err
		}
		return payload(sc)
	}); err != nil {
		return err
	}
	return nil
}

func parseFilter(filter map[string]interface{}) map[string]interface{} {
	if filter["_id"] != nil {
		id := fmt.Sprintf("%v", filter["_id"])
		filter["_id"] = parseStringToMongo(&id)
	}
	return filter
}

func createCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

func parsePayload[T MongoModels](payload T) *T {
	byteA := dataToByteA(payload)
	payload_map := *byteAToData[map[string]interface{}](byteA)
	if payload_map["Id"] == "000000000000000000000000" {
		payload_map["Id"] = primitive.NewObjectID()
	} else {
		payload_map["Id"] = parseStringToMongo(payload_map["Id"].(*string))
	}
	return byteAToData[T](dataToByteA(payload_map))
}

func parseMultiple[T MongoModels](payload []T) []*T {
	var result []*T
	for _, data := range payload {
		result = append(result, parsePayload(data))
	}
	return result
}

func byteAToData[T interface{}](payload []byte) *T {
	var data T
	json.Unmarshal(payload, &data)
	return &data
}

func dataToByteA(payload interface{}) []byte {
	data, _ := json.Marshal(payload)
	return data
}

func parseStringToMongo(id *string) primitive.ObjectID {
	objId, _ := primitive.ObjectIDFromHex(*id)
	return objId
}

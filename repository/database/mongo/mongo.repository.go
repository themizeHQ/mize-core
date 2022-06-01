package repository

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	appModel "mize.app/app/application/models"
	"mize.app/app/user/models"
	"mize.app/app_errors"
)

type MongoModels interface {
	user.User | appModel.Application
}

type MongoRepository[T MongoModels] struct {
	Model   *mongo.Collection
	Payload interface{}
}

func (repo *MongoRepository[T]) CreateOne(ctx *gin.Context, payload *T, opts ...*options.InsertOneOptions) mongo.InsertOneResult {
	c, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	result, err := repo.Model.InsertOne(c, payload, opts...)

	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
	}

	defer func() {
		cancel()
	}()

	return *result
}

func (repo *MongoRepository[T]) FindOneByFilter(ctx *gin.Context, filter map[string]interface{}, opts ...*options.FindOneOptions) *user.User {
	filter = parseFilter(ctx, filter)
	c, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer func() {
		cancel()
	}()

	var resultDecoded user.User
	cursor := repo.Model.FindOne(c, filter, opts...)
	err := cursor.Decode(&resultDecoded)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil
		}

		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return nil
	}

	return &resultDecoded
}

func (repo *MongoRepository[T]) CountDocs(ctx *gin.Context, filter map[string]interface{}, opts ...*options.CountOptions) int64 {
	filter = parseFilter(ctx, filter)
	c, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer func() {
		cancel()
	}()

	count, err := repo.Model.CountDocuments(c, filter, opts...)
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return 0
	}
	return count
}

func parseFilter(ctx *gin.Context, filter map[string]interface{}) map[string]interface{} {
	if filter["_id"] != nil {
		id := fmt.Sprintf("%v", filter["_id"])
		objId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
			return filter
		}
		filter["_id"] = objId
	}
	return filter
}

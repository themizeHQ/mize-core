package repository

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/user/models"
	"mize.app/app_errors"
)

type MongoModels interface {
	user.User
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

func (repo *MongoRepository[T]) FindOneByFilter(ctx *gin.Context, filter interface{}, opts ...*options.FindOneOptions) *user.User {
	c, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer func() {
		cancel()
	}()

	var resultDecoded user.User
	cursor := repo.Model.FindOne(c, filter, opts...)
	err := cursor.Decode(&resultDecoded)
	fmt.Println(resultDecoded)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil
		}

		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return nil
	}

	return &resultDecoded
}

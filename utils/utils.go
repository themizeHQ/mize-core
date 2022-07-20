package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mize.app/app_errors"
)

func HexToMongoId(ctx *gin.Context, id string) *primitive.ObjectID {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return &objId
	}
	return &objId
}

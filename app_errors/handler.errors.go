package app_errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mize.app/server_response"
)

type MizeErrors interface {
	RequestError
}

func ErrorHandler(ctx *gin.Context, err error) {
	ctx.Abort()
	server_response.Response(ctx, http.StatusInternalServerError, err. Error(), false, nil)
}

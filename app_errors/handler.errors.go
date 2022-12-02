package app_errors

import (
	"github.com/gin-gonic/gin"
	"mize.app/server_response"
)

type MizeErrors interface {
	RequestError
}

func ErrorHandler(ctx *gin.Context, err RequestError) {
	ctx.Abort()
	server_response.Response(ctx, err.StatusCode, err.Error(), false, nil)
}

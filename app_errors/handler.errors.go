package app_errors

import (
	"github.com/gin-gonic/gin"
	"mize.app/server_response"
)

type MizeErrors interface {
	RequestError
}

func ErrorHandler(ctx *gin.Context, err error, code int) {
	ctx.Abort()
	server_response.Response(ctx, code, err.Error(), false, nil)
}

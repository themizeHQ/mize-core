package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"mize.app/app/conversation/types"
	"mize.app/app_errors"
	"mize.app/server_response"

	reactionUseCases "mize.app/app/conversation/usecases/reaction"
)

func CreateReaction(ctx *gin.Context) {
	var data types.Reaction
	if err := ctx.ShouldBind(&data); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the valid payload"), StatusCode: http.StatusBadRequest})
		return
	}
	err := reactionUseCases.CreateReactionUseCase(ctx, data, ctx.Params.ByName("channel") == "true")
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "reacted", true, nil)
}

func RemoveReaction(ctx *gin.Context) {
	var data types.RemoveReaction
	if err := ctx.ShouldBind(&data); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the valid payload"), StatusCode: http.StatusBadRequest})
		return
	}
	err := reactionUseCases.RemoveReactionUseCase(ctx, data, ctx.Params.ByName("channel") == "true")
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusOK, "reacted", true, nil)
}

package controllers

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"mize.app/app/conversation/types"
	convUsecases "mize.app/app/conversation/usecases/conversation"
	"mize.app/app_errors"
	"mize.app/server_response"
	"mize.app/utils"
)

func StartConversation(ctx *gin.Context) {
	var payload types.CreateConv
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in message data"), StatusCode: http.StatusBadRequest})
		return
	}
	convId, reciepientName, recipientImage, err := convUsecases.CreateConversationUseCase(ctx, &payload)
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	chan1 := make(chan error)
	wg.Add(1)
	go func(e chan error) {
		defer func() {
			wg.Done()
		}()
		profileImage := ctx.GetString("ProfileImage")
		_, err := convUsecases.CreateConversationMemberUseCase(ctx, &payload, convId, *reciepientName, payload.ReciepientId.Hex(), *utils.HexToMongoId(ctx, ctx.GetString("UserId")), &profileImage)
		e <- err
	}(chan1)

	wg.Add(1)
	chan2 := make(chan error)
	go func(e chan error) {
		defer func() {
			wg.Done()
		}()
		_, err := convUsecases.CreateConversationMemberUseCase(ctx, &payload, convId, ctx.GetString("Username"), ctx.GetString("UserId"), payload.ReciepientId, recipientImage)
		e <- err
	}(chan2)

	err1 := <-chan1
	err2 := <-chan2
	if err1 != nil || err2 != nil {
		return
	}
	wg.Wait()
	server_response.Response(ctx, http.StatusOK, "conversation created", true, nil)
}

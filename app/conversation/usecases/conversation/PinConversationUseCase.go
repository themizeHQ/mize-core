package conversation

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mize.app/app/conversation/repository"
	"mize.app/app_errors"
	"mize.app/utils"
)

func PinConversationUseCase(ctx *gin.Context, id *string) error {
	conversationMemberRepo := repository.GetConversationMemberRepo()
	exists, err := conversationMemberRepo.CountDocs(map[string]interface{}{
		"userId": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"_id":    *id,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong"), StatusCode: http.StatusInternalServerError})
		return err
	}
	if exists == 0 {
		err = errors.New("this conversation does not exist")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return err
	}
	pinnedCount, err := conversationMemberRepo.CountDocs(map[string]interface{}{
		"userId": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"pinned": true,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong"), StatusCode: http.StatusInternalServerError})
		return err
	}
	if pinnedCount >= 5 {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you can pin only 5 DMs"), StatusCode: http.StatusUnauthorized})
		return err
	}
	success, err := conversationMemberRepo.UpdatePartialById(ctx, *id, map[string]interface{}{
		"pinned": true,
	})
	if err != nil || !success {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong"), StatusCode: http.StatusInternalServerError})
		return err
	}
	return nil
}

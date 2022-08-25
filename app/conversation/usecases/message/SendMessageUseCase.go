package messages

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mize.app/app/conversation/models"
	"mize.app/app/conversation/repository"
	"mize.app/app_errors"
	"mize.app/emitter"
	"mize.app/utils"
)

func SendMessageUseCase(ctx *gin.Context, payload models.Message) error {
	messageRepository := repository.GetMessageRepo()
	var replyTo string
	if payload.ReplyTo != nil {
		exist, err := messageRepository.CountDocs(map[string]interface{}{
			"_id": payload.ReplyTo.Hex(),
		})
		if err != nil {
			err = errors.New("an error occured")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
			return err
		}
		if exist != 1 {
			err = errors.New("message has been deleted")
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
			return err
		}
		replyTo = payload.ReplyTo.Hex()
	}
	payload.WorkspaceId = *utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
	payload.From = *utils.HexToMongoId(ctx, ctx.GetString("UserId"))
	payload.Username = ctx.GetString("Username")
	_, err := messageRepository.CreateOne(payload)
	if err != nil {
		err = errors.New("message could not be sent")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return err
	}
	emitter.Emitter.Emit(emitter.Events.MESSAGES_EVENTS.MESSAGE_SENT, map[string]interface{}{
		"time":     time.Now(),
		"from":     payload.From.Hex(),
		"to":       payload.To.Hex(),
		"text":     payload.Text,
		"replyTo":  replyTo,
		"username": payload.Username,
	})
	return nil
}

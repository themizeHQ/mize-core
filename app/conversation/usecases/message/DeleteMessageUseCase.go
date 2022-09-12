package messages

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	conversationRepository "mize.app/app/conversation/repository"
	types "mize.app/app/conversation/types"
	"mize.app/app_errors"

	channelRepository "mize.app/app/workspace/repository"
)

func DeleteMessageUseCase(ctx *gin.Context, toDel []types.DeleteMessage) error {
	var wg sync.WaitGroup
	channelMemberRepo := channelRepository.GetChannelMemberRepo()

	chan1 := make(chan app_errors.RequestError)
	var err error
	for _, m := range toDel {
		wg.Add(1)
		go func(m types.DeleteMessage, e chan app_errors.RequestError) {
			defer func() {
				wg.Done()
			}()
			exists, err := channelMemberRepo.CountDocs(map[string]interface{}{
				"channelId":   m.To,
				"workspaceId": ctx.GetString("Workspace"),
				"userId":      ctx.GetString("UserId"),
			})
			if err != nil {
				e <- app_errors.RequestError{Err: errors.New("could not delete message"), StatusCode: http.StatusInternalServerError}
				return
			}
			if exists != 1 {
				e <- app_errors.RequestError{Err: errors.New("you are not a member of this channel"), StatusCode: http.StatusUnauthorized}
				return
			}
			fmt.Println(exists)
			messageRepo := conversationRepository.GetMessageRepo()
			messageRepo.DeleteOne(map[string]interface{}{
				"_id": m.MessageId,
				"to":  m.To,
			})
			e <- app_errors.RequestError{}
		}(m, chan1)
		err1 := <-chan1
		if err1.Err != nil && err != nil {
			fmt.Println(err1.Err)
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you are not a member of this channel"), StatusCode: http.StatusUnauthorized})
			err = err1.Err
		}
	}
	wg.Wait()
	return err
}

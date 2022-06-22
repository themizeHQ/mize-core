package workspace

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	workspaceModel "mize.app/app/workspace/models"
	workspace_invite "mize.app/app/workspace/usecases/workspace_invite"
	"mize.app/app_errors"
	"mize.app/emails"
)

var wg sync.WaitGroup

func SendInvitesUseCase(ctx *gin.Context, user_emails []string, workspace_name string) error {
	var err error
	for _, email := range user_emails {
		wg.Add(1)
		go func(e string) {
			defer func() {
				wg.Done()
			}()
			result := emails.SendEmail(e, fmt.Sprintf("You are invited to join the workspace, %s", workspace_name),
				"workspace_invite", map[string]string{"WORKSPACE_NAME": workspace_name, "LINK": "should be a url here"})
			workspace_invite.CreateWorkspaceInviteUseCase(ctx, workspaceModel.WorkspaceInvite{Email: e, Success: result})
			if !result {
				err = errors.New("an error occured while sending invites")
				app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
			}
		}(email)
	}
	wg.Wait()
	return err
}

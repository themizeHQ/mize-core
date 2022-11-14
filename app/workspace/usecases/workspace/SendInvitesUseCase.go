package workspace

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"mize.app/app/workspace/usecases/workspace_invite"
	"mize.app/app_errors"
	"mize.app/emails"
	"mize.app/utils"
)

func SendInvitesUseCase(ctx *gin.Context, user_emails []string) error {
	failed := 0
	var wg sync.WaitGroup
	for _, email := range user_emails {
		wg.Add(1)
		go func(e string) {
			defer func() {
				wg.Done()
			}()
			success := emails.SendEmail(e, fmt.Sprintf("You are invited to join the workspace, %s", ctx.GetString("WorkspaceName")),
				"workspace_invite", map[string]string{"WORKSPACE_NAME": ctx.GetString("WorkspaceName"), "LINK": fmt.Sprintf("https://mize.app?%s&?%s", ctx.GetString("Workspace"), e)})
			er := workspace_invite.CreateWorkspaceInviteUseCase(ctx, map[string]interface{}{
				"email": e, "workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
			}, map[string]interface{}{"email": e, "success": success, "workspaceName": ctx.GetString("WorkspaceName"), "workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace"))})
			if !success || er != nil {
				failed++
			}
		}(email)
	}
	wg.Wait()
	if failed != 0 {
		err := fmt.Errorf("%d invites had problems. examine the info provided on the admin dashboard", failed)
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	return nil
}

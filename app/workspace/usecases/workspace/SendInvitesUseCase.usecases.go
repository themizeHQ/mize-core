package workspace

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	workspaceInvite "mize.app/app/workspace/usecases/workspace_invite"
	"mize.app/app_errors"
	"mize.app/emails"
)

func SendInvitesUseCase(ctx *gin.Context, user_emails []string, workspace_id string) error {
	var workspaceRepoInstance = workspaceRepo.GetWorkspaceRepo()
	workspace, err := workspaceRepoInstance.FindById(workspace_id)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("workspace does not exist"), StatusCode: http.StatusNotFound})
		return err
	}
	has_access := false
	if workspace.CreatedBy == ctx.GetString("UserId") {
		has_access = true
	} else {
		for _, admin := range workspace.Admins {
			if admin == ctx.GetString("UserId") {
				has_access = true
				break
			}
		}
	}
	if !has_access {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("only admins can send invites to workspaces"), StatusCode: http.StatusUnauthorized})
		return err
	}
	failed := 0
	var wg sync.WaitGroup
	for _, email := range user_emails {
		ch_mail_sent := make(chan bool, 1)
		wg.Add(1)
		go func(e string) {
			defer func() {
				wg.Done()
			}()
			ch_mail_sent <- emails.SendEmail(e, fmt.Sprintf("You are invited to join the workspace, %s", workspace.Name),
				"workspace_invite", map[string]string{"WORKSPACE_NAME": workspace.Name, "LINK": "should be a url here"})
			if !<-ch_mail_sent {
				failed++
			}
			close(ch_mail_sent)
		}(email)

		wg.Add(1)
		go func(e string) {
			defer func() {
				wg.Done()
			}()
			workspaceInvite.CreateWorkspaceInviteUseCase(ctx, workspaceModel.WorkspaceInvite{Email: e, Success: <-ch_mail_sent, Workspace: workspace_id})
		}(email)
	}
	wg.Wait()
	if failed != 0 {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: fmt.Errorf("%d invites failed to send", failed), StatusCode: http.StatusBadRequest})
		return err
	}
	return err
}

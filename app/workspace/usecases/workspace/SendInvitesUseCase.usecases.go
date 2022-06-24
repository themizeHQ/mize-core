package workspace

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	workspace_invite "mize.app/app/workspace/usecases/workspace_invite"
	"mize.app/app_errors"
	"mize.app/emails"
)

var wg sync.WaitGroup

func SendInvitesUseCase(ctx *gin.Context, user_emails []string, workspace_id string) error {
	workspace := workspaceRepo.WorkspaceRepository.FindOneByFilter(ctx, map[string]interface{}{
		"_id": workspace_id,
	})
	var err error
	if workspace == nil {
		err = errors.New("workspace does not exist")
		app_errors.ErrorHandler(ctx, err, http.StatusNotFound)
		return err
	}
	has_access := false
	if workspace.CreatedBy.Hex() == ctx.GetString("UserId") {
		has_access = true
	} else {
		for _, admin := range workspace.Admins {
			if admin.Hex() == ctx.GetString("UserId") {
				has_access = true
				break
			}
		}
	}
	if !has_access {
		err = errors.New("only admins can send invites for workspaces")
		app_errors.ErrorHandler(ctx, err, http.StatusUnauthorized)
		return err
	}
	for _, email := range user_emails {
		wg.Add(1)
		go func(e string) {
			defer func() {
				wg.Done()
			}()
			result := emails.SendEmail(e, fmt.Sprintf("You are invited to join the workspace, %s", workspace.Name),
				"workspace_invite", map[string]string{"WORKSPACE_NAME": workspace.Name, "LINK": "should be a url here"})
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

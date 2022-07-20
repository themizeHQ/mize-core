package workspace

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	// workspaceModel "mize.app/app/workspace/models"
	workspaceRepo "mize.app/app/workspace/repository"
	workspaceInvite "mize.app/app/workspace/usecases/workspace_invite"
	"mize.app/app_errors"
	"mize.app/emails"
)

func SendInvitesUseCase(ctx *gin.Context, user_emails []string) error {
	var workspaceRepoInstance = workspaceRepo.GetWorkspaceRepo()
	var workspaceMemberRepoInstance = workspaceRepo.GetWorkspaceMember()
	workspace, err := workspaceRepoInstance.FindById(ctx.GetString("Workspace"))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("workspace does not exist"), StatusCode: http.StatusNotFound})
		return err
	}
	workspace_member, err := workspaceMemberRepoInstance.FindOneByFilter(map[string]interface{}{"workspaceId": ctx.GetString("Workspace"),
		"userId": ctx.GetString("UserId")})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("workspace does not exist"), StatusCode: http.StatusNotFound})
		return err
	}
	if !workspace_member.Admin {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("only admins can send invites to workspaces"), StatusCode: http.StatusUnauthorized})
		return err
	}
	failed := 0
	var wg sync.WaitGroup
	for _, email := range user_emails {
		wg.Add(1)
		go func(e string) {
			defer func() {
				wg.Done()
			}()
			success := emails.SendEmail(e, fmt.Sprintf("You are invited to join the workspace, %s", workspace.Name),
				"workspace_invite", map[string]string{"WORKSPACE_NAME": workspace.Name, "LINK": fmt.Sprintf("https://mize.app?%s&?%s", ctx.GetString("Workspace"), e)})
			er := workspaceInvite.CreateWorkspaceInviteUseCase(ctx, map[string]interface{}{"email": e, "workspaceId": workspace}, map[string]interface{}{"email": e, "success": success, "workspaceId": ctx.GetString("Workspace")})
			if !success || er != nil {
				failed++
			}
		}(email)
	}
	wg.Wait()
	if failed != 0 {
		err = fmt.Errorf("%d invites had problems. examine the info provided on the admin dashboard", failed)
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return err
	}
	return err
}

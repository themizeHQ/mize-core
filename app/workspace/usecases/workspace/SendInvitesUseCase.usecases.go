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
	workspace, err := workspaceRepoInstance.FindById(ctx.GetString("Workspace"))
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
		wg.Add(1)
		go func(e string) {
			defer func() {
				wg.Done()
			}()
			success := emails.SendEmail(e, fmt.Sprintf("You are invited to join the workspace, %s", workspace.Name),
				"workspace_invite", map[string]string{"WORKSPACE_NAME": workspace.Name, "LINK": fmt.Sprintf("https://mize.app?%s&?%s", ctx.GetString("Workspace"), e)})
			er := workspaceInvite.CreateWorkspaceInviteUseCase(ctx, map[string]interface{}{"email": e, "workspace": ctx.GetString("Workspace")}, map[string]interface{}{"email": e, "success": success, "workspace": ctx.GetString("Workspace")})
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

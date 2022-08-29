package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/media"
	"mize.app/app/workspace/models"
	workspaceUseCases "mize.app/app/workspace/usecases/workspace"
	workspaceMemberUseCases "mize.app/app/workspace/usecases/workspace_member"
	"mize.app/app_errors"
	mediaConstants "mize.app/constants/media"
	"mize.app/server_response"
	"mize.app/utils"
)

func CreateWorkspace(ctx *gin.Context) {
	var payload models.Workspace
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json value"), StatusCode: http.StatusBadRequest})
		return
	}
	name, id, err := workspaceUseCases.CreateWorkspaceUseCase(ctx, payload)
	if err != nil {
		return
	}
	member_id, err := workspaceMemberUseCases.CreateWorkspaceMemberUseCase(ctx, id, name, true)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "workspace successfully created.", true, map[string]interface{}{
		"workspace_id":        id,
		"workspace_member_id": *member_id,
	})
}

func UpdatWorkspaceProfileImage(ctx *gin.Context) {
	file, fileHeader, err := ctx.Request.FormFile("media")
	defer func() {
		file.Close()
	}()
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	uploadRepo := media.GetUploadRepo()
	profileImageUpload, err := uploadRepo.FindOneByFilter(map[string]interface{}{
		"uploadBy": *utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		"type":     mediaConstants.WORKSPACE_IMAGE,
	}, options.FindOne().SetProjection(map[string]interface{}{
		"publicId": 1,
	}))
	if err != nil {
		err := errors.New("could not update profile image")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	var data *media.Upload
	if profileImageUpload == nil {
		data, err = media.UploadToCloudinary(ctx, file, "/core/workspace-images", nil)
		if err != nil {
			return
		}
	} else {
		data, err = media.UploadToCloudinary(ctx, file, "/core/workspace-images", &profileImageUpload.PublicID)
		if err != nil {
			return
		}
	}
	data.Type = mediaConstants.WORKSPACE_IMAGE
	data.Format = strings.Split(fileHeader.Filename, ".")[len(strings.Split(fileHeader.Filename, "."))-1]
	data.UploadBy = *utils.HexToMongoId(ctx, ctx.GetString("Workspace"))

	if profileImageUpload == nil {
		err = workspaceUseCases.UploadWorkspaceImage(ctx, data)
		if err != nil {
			return
		}
	} else {
		err = workspaceUseCases.UpdateWorkspaceImage(ctx, data.Bytes)
		if err != nil {
			return
		}
	}
	server_response.Response(ctx, http.StatusCreated, "upload success", true, nil)
}

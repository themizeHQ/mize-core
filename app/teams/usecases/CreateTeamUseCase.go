package usecases

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"mize.app/app/media"
	"mize.app/app/teams/models"
	"mize.app/app/teams/repository"
	"mize.app/app_errors"
	mediaConstants "mize.app/constants/media"
	"mize.app/utils"
)

func CreateTeamUseCase(ctx *gin.Context, payload *models.Team, image multipart.File, format string, fileName string) error {
	teamRepo := repository.GetTeamRepo()
	payload.WorkspaceId = *utils.HexToMongoId(ctx, ctx.GetString("Workspace"))
	var upload *media.Upload
	var err error
	if image != nil {
		upload, err = media.UploadToCloudinary(ctx, image, "/core/teams", nil)
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not create team"), StatusCode: http.StatusInternalServerError})
			return err
		}
	}
	if upload != nil {
		payload.ProfileImage = &upload.Url
	}
	team, err := teamRepo.CreateOne(*payload)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not create team"), StatusCode: http.StatusInternalServerError})
		return err
	}
	if upload != nil {
		uploadRepo := media.GetUploadRepo()
		upload.UploadBy = team.Id
		upload.Type = mediaConstants.PROFILE_IMAGE
		upload.Format = format
		upload.FileName = fileName
		_, err = uploadRepo.CreateOne(*upload)
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not create team"), StatusCode: http.StatusInternalServerError})
			return err
		}
	}
	return nil
}

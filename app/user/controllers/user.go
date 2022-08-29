package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mize.app/app/media"
	"mize.app/app/user/models"
	userRepo "mize.app/app/user/repository"
	userUseCases "mize.app/app/user/usecases/user"
	"mize.app/app_errors"
	mediaConstants "mize.app/constants/media"
	"mize.app/server_response"
	"mize.app/utils"
)

func FetchProfile(ctx *gin.Context) {
	userRepo := userRepo.GetUserRepo()
	profile, err := userRepo.FindById(ctx.GetString("UserId"), options.FindOne().SetProjection(map[string]int{
		"password": 0,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("profile not found. please contact support"), StatusCode: http.StatusInternalServerError})
		return
	}
	profile.Password = ""
	server_response.Response(ctx, http.StatusOK, "profile fetched", true, profile)
}

func FetchUsersProfile(ctx *gin.Context) {
	userRepo := userRepo.GetUserRepo()
	id := ctx.Params.ByName("id")
	if id == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in the user id"), StatusCode: http.StatusBadRequest})
		return
	}
	profile, err := userRepo.FindById(id, options.FindOne().SetProjection(map[string]int{
		"password": 0,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not find user"), StatusCode: http.StatusInternalServerError})
		return
	}
	profile.Password = ""
	server_response.Response(ctx, http.StatusOK, "profile fetched", true, profile)
}

func UpdateUserData(ctx *gin.Context) {
	var payload models.UpdateUser
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json with valid fields"), StatusCode: http.StatusBadRequest})
		return
	}
	payload.FirstName = strings.TrimSpace(payload.FirstName)
	err := payload.ValidateUpdate()
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	success := userUseCases.UpdateUserUseCase(ctx, payload)
	if !success {
		return
	}
	server_response.Response(ctx, http.StatusOK, "profile updated", success, nil)
}

func UpdateProfileImage(ctx *gin.Context) {
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
		"uploadBy": utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"type":     mediaConstants.PROFILE_IMAGE,
	}, options.FindOne().SetProjection(map[string]interface{}{
		"publicId": 1,
	}))
	if err != nil {
		fmt.Println(err)
		err := errors.New("could not update profile image")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	if profileImageUpload == nil {
		data, err := media.UploadToCloudinary(ctx, file, "/core/profile-images", nil)
		if err != nil {
			return
		}
		data.Type = mediaConstants.PROFILE_IMAGE
		data.Format = strings.Split(fileHeader.Filename, ".")[len(strings.Split(fileHeader.Filename, "."))-1]
		err = userUseCases.UploadProfileImage(ctx, data)
		if err != nil {
			return
		}
	} else {
		data, err := media.UploadToCloudinary(ctx, file, "/core/profile-images", &profileImageUpload.PublicID)
		if err != nil {
			return
		}
		data.Type = mediaConstants.PROFILE_IMAGE
		data.Format = strings.Split(fileHeader.Filename, ".")[len(strings.Split(fileHeader.Filename, "."))-1]
		err = userUseCases.UpdateProfileImage(ctx, data.Bytes)
		if err != nil {
			return
		}
	}
	server_response.Response(ctx, http.StatusCreated, "upload success", true, nil)
}

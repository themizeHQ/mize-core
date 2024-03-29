package controllers

import (
	"errors"
	"fmt"

	// "fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mize.app/app/media"
	"mize.app/app/user/models"
	userRepo "mize.app/app/user/repository"
	"mize.app/app/user/types"
	userUseCases "mize.app/app/user/usecases/user"
	"mize.app/app_errors"
	"mize.app/authentication"
	mediaConstants "mize.app/constants/media"
	user_constants "mize.app/constants/user"
	"mize.app/emitter"
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
	if profile == nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("user does not exist"), StatusCode: http.StatusNotFound})
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
		"uploadBy": *utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"type":     mediaConstants.PROFILE_IMAGE,
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
		userId := ctx.GetString("UserId")
		data, err = media.UploadToCloudinary(ctx, file, fmt.Sprintf("/core/profile-images/%s", userId), nil)
		if err != nil {
			return
		}
	} else {
		data, err = media.UploadToCloudinary(ctx, file, "", &profileImageUpload.PublicID)
		if err != nil {
			return
		}
	}

	data.Type = mediaConstants.PROFILE_IMAGE
	data.Format = strings.Split(fileHeader.Filename, ".")[len(strings.Split(fileHeader.Filename, "."))-1]
	data.UploadBy = *utils.HexToMongoId(ctx, ctx.GetString("UserId"))

	if profileImageUpload == nil {
		err = userUseCases.UploadProfileImage(ctx, data)
		if err != nil {
			return
		}
	} else {
		err = userUseCases.UpdateProfileImage(ctx, data.Bytes)
		if err != nil {
			return
		}
	}
	server_response.Response(ctx, http.StatusCreated, "upload success", true, nil)
}

func UpdatePhone(ctx *gin.Context) {
	var payload struct {
		Phone string
	}
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a valid phone number with country code"), StatusCode: http.StatusBadRequest})
		return
	}
	userRepo := userRepo.GetUserRepo()
	profile, err := userRepo.CountDocs(map[string]interface{}{
		"phone": payload.Phone,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong while adding number"), StatusCode: http.StatusInternalServerError})
		return
	}
	fmt.Println(profile)
	if profile != 0 {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("user with mobile number already exists"), StatusCode: http.StatusConflict})
		return
	}
	otp, err := authentication.GenerateOTP(6)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not generate otp"), StatusCode: http.StatusBadRequest})
		return
	}
	authentication.SaveOTP(ctx, payload.Phone, otp, 5*time.Minute)
	emitter.Emitter.Emit(emitter.Events.SMS_EVENTS.SMS_SENT, map[string]interface{}{
		"to":      payload.Phone,
		"message": fmt.Sprintf("Your Mize otp is %s \n This OTP is valid for 5 minutes.", otp),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not send otp"), StatusCode: http.StatusBadRequest})
		return
	}
	server_response.Response(ctx, http.StatusCreated, "otp sent", true, nil)
}

func FetchUsersByPhoneNumber(ctx *gin.Context) {
	var filter types.PhoneNumberFilter
	if err := ctx.ShouldBind(&filter); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass an array of mobile numbers"), StatusCode: http.StatusBadRequest})
		return
	}
	if len(filter.Phone) <= 0 {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in at least 1 mobile number"), StatusCode: http.StatusBadRequest})
		return
	}
	userRepo := userRepo.GetUserRepo()
	profile, err := userRepo.FindManyStripped(map[string]interface{}{
		"phone": map[string]interface{}{
			"$in": filter.Phone,
		},
		"discoverability": user_constants.DISCOVERABILITY_PHONE,
	}, options.Find().SetProjection(map[string]int{
		"firstName":    1,
		"lastName":     1,
		"profileImage": 1,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not find users"), StatusCode: http.StatusInternalServerError})
		return
	}
	server_response.Response(ctx, http.StatusOK, "users found", true, profile)
}

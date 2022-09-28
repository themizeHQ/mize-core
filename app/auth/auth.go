package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/user/models"
	"mize.app/app/user/repository"
	userUseCases "mize.app/app/user/usecases/user"
	workspaceRepository "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/constants/channel"
	user_constants "mize.app/constants/user"
	"mize.app/cryptography"
	"mize.app/utils"

	"mize.app/calls/azure"
	"mize.app/emitter"
	"mize.app/realtime"
	"mize.app/repository/database/redis"
	"mize.app/server_response"
)

func CacheUserUseCase(ctx *gin.Context) {
	var userRepoInstance = repository.GetUserRepo()

	var payload models.User
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	payload.Status = user_constants.AVAILABLE
	if payload.Language == "" {
		payload.Language = "english"
	}
	payload.Discoverability = []user_constants.UserDiscoverability{user_constants.DISCOVERABILITY_EMAIL, user_constants.DISCOVERABILITY_USERNAME, user_constants.DISCOVERABILITY_PHONE}
	payload.RunHooks()
	emailExists, err := userRepoInstance.CountDocs(map[string]interface{}{"email": payload.Email})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return
	}
	if emailExists != 0 {
		err := errors.New("user with email already exists")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
		return
	}
	usernameExists, err := userRepoInstance.CountDocs(map[string]interface{}{"userName": payload.UserName})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return
	}
	if usernameExists != 0 {
		err := errors.New("user with username already exists")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusConflict})
		return
	}
	result := redis.RedisRepo.CreateEntry(ctx, fmt.Sprintf("%s-user", payload.Email), payload, 20*time.Minute)
	if !result {
		err := errors.New("something went wrong while creating user")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return
	}
	otp, err := authentication.GenerateOTP(6)
	if err != nil {
		server_response.Response(ctx, http.StatusInternalServerError, "something went wrong while creating user", false, nil)
		return
	}
	authentication.SaveOTP(ctx, payload.Email, otp, 5*time.Minute)
	emitter.Emitter.Emit(emitter.Events.AUTH_EVENTS.USER_CREATED, map[string]interface{}{"email": payload.Email, "otp": strings.Split(otp, "")})
	server_response.Response(ctx, http.StatusCreated, "user created successfuly", true, nil)
}

func VerifyAccountUseCase(ctx *gin.Context) {
	var userRepoInstance = repository.GetUserRepo()

	type VerifyData struct {
		Otp   string
		Email string
	}
	var payload VerifyData
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	valid, err := authentication.VerifyOTP(ctx, fmt.Sprintf("%s-otp", payload.Email), payload.Otp)
	if err != nil {
		return
	}
	if !valid {
		server_response.Response(ctx, http.StatusUnauthorized, "wrong otp provided", false, nil)
		return
	}
	cached_user := redis.RedisRepo.FindOne(ctx, fmt.Sprintf("%s-user", payload.Email))
	if cached_user == nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("this user does not exist"), StatusCode: http.StatusNotFound})
		return
	}
	var data models.User
	json.Unmarshal([]byte(*cached_user), &data)
	acsData, err := azure.GenerateUserAndToken()
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return
	}
	data.ACSUserId = acsData.User.CommunicationUserId
	response, err := userRepoInstance.CreateOne(data)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong while creating user"), StatusCode: http.StatusNotFound})
		return
	}
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-user", payload.Email))
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-otp", payload.Email))
	err = authentication.GenerateRefreshToken(ctx, response.Id.Hex(), payload.Email, data.UserName, data.FirstName, data.Language, acsData.User.CommunicationUserId)
	if err != nil {
		return
	}
	err = authentication.GenerateAccessToken(ctx, response.Id.Hex(), payload.Email, data.UserName, data.FirstName, data.LastName, nil, acsData.User.CommunicationUserId)
	if err != nil {
		return
	}
	emitter.Emitter.Emit(emitter.Events.AUTH_EVENTS.USER_VERIFIED, map[string]string{"email": payload.Email})
	response.Password = ""
	response.ACSUserId = ""
	server_response.Response(ctx, http.StatusCreated, "account verified", true, map[string]interface{}{
		"user": response,
	})
}

func LoginUser(ctx *gin.Context) {
	var userRepoInstance = repository.GetUserRepo()

	type LoginDetails struct {
		Password string
		Account  string
	}
	var payload LoginDetails
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json"), StatusCode: http.StatusBadRequest})
		return
	}
	var profile *models.User
	if _, err := mail.ParseAddress(payload.Account); err == nil {
		p, err := repository.UserRepository.FindOneByFilter(map[string]interface{}{"email": payload.Account})
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
			return
		}
		profile = p
	} else {
		p, err := userRepoInstance.FindOneByFilter(map[string]interface{}{"userName": payload.Account})
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
			return
		}
		profile = p
	}
	if profile == nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("user does not exist"), StatusCode: http.StatusNotFound})
		return
	}
	success := cryptography.VerifyData(profile.Password, payload.Password)
	if !success {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("incorrect password"), StatusCode: http.StatusUnauthorized})
		return
	}
	err := authentication.GenerateRefreshToken(ctx, profile.Id.Hex(), profile.Email, profile.UserName, profile.FirstName, profile.LastName, profile.ACSUserId)
	if err != nil {
		return
	}
	err = authentication.GenerateAccessToken(ctx, profile.Id.Hex(), profile.Email, profile.UserName, profile.FirstName, profile.LastName, nil, profile.ACSUserId)
	if err != nil {
		return
	}
	profile.Password = ""
	profile.ACSUserId = ""
	server_response.Response(ctx, http.StatusCreated, "login successful", true, map[string]interface{}{
		"user": profile,
	})
}

func GenerateAcsToken(ctx *gin.Context) {
	id := ctx.GetString("ACSUserId")
	acsDetails, err := azure.RefreshToken(&id)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return
	}
	server_response.Response(ctx, http.StatusCreated, "token generated", true, map[string]interface{}{
		"acsDetails": map[string]interface{}{
			"token":     acsDetails.Token,
			"expiresOn": acsDetails.ExpiresOn,
		},
	})
}

func UpdateLoggedInUsersPassword(ctx *gin.Context) {
	var payload struct {
		CurrentPassword string
		NewPassword     string
	}
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in both your current password and new password"), StatusCode: http.StatusBadRequest})
		return
	}
	userRepoInstance := repository.GetUserRepo()
	user, err := userRepoInstance.FindById(ctx.GetString("UserId"), options.FindOne().SetProjection(map[string]int{
		"password": 1,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("user does not exist"), StatusCode: http.StatusNotFound})
		return
	}
	success := cryptography.VerifyData(user.Password, payload.CurrentPassword)
	if !success {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("incorrect password"), StatusCode: http.StatusUnauthorized})
		return
	}
	user.Password = payload.NewPassword
	err = user.ValidatePassword()
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	success, err = userRepoInstance.UpdatePartialById(ctx, ctx.GetString("UserId"), map[string]interface{}{
		"password": string(cryptography.HashString(user.Password, nil)),
	})
	if !success || err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong while updating password"), StatusCode: http.StatusInternalServerError})
	}
	server_response.Response(ctx, http.StatusOK, "password change successful", true, nil)
}

func GenerateAccessTokenFromRefresh(ctx *gin.Context) {
	refresh_token := ctx.GetHeader("Authorization")
	if refresh_token == "" {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("no auth tokens provided"), StatusCode: http.StatusUnauthorized})
		return
	}

	refresh_token = strings.Split(refresh_token, " ")[1]

	valid_refresh_token := authentication.DecodeAuthToken(ctx, refresh_token)
	if !valid_refresh_token.Valid {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("invalid refresh token used"), StatusCode: http.StatusUnauthorized})
		return
	}

	refresh_token_claims := valid_refresh_token.Claims.(jwt.MapClaims)
	if refresh_token_claims["Type"].(string) != string(authentication.REFRESH_TOKEN) {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("invalid refresh token used"), StatusCode: http.StatusUnauthorized})
		return
	}
	workspace := ctx.Query("workspace_id")
	if workspace == "" {
		authentication.GenerateAccessToken(ctx, refresh_token_claims["UserId"].(string),
			refresh_token_claims["Email"].(string), refresh_token_claims["Username"].(string), refresh_token_claims["Firstname"].(string), refresh_token_claims["Lastname"].(string), nil, refresh_token_claims["ACSUserId"].(string))
		server_response.Response(ctx, http.StatusCreated, "token generated", true, nil)
		return
	}
	err := authentication.GenerateAccessToken(ctx, refresh_token_claims["UserId"].(string),
		refresh_token_claims["Email"].(string), refresh_token_claims["Username"].(string), refresh_token_claims["Firstname"].(string), refresh_token_claims["Lastname"].(string), &workspace, refresh_token_claims["ACSUserId"].(string))
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "token generated", true, nil)
}

func GenerateCentrifugoToken(ctx *gin.Context) {
	token, err := authentication.GenerateCentrifugoAuthToken(ctx, jwt.StandardClaims{
		Subject:   ctx.GetString("UserId"),
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(1)).Unix(),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not generate token"), StatusCode: http.StatusInternalServerError})
		return
	}
	if ctx.Query("channels") == "true" {
		server_response.Response(ctx, http.StatusCreated, "token generated", true, map[string]interface{}{
			"token": token,
			"channels": map[string]interface{}{
				"default_channels": realtime.DefaultChannels,
			},
		})
	} else {
		server_response.Response(ctx, http.StatusCreated, "token generated", true, token)
	}
}

func ResendOtp(ctx *gin.Context) {
	email := ctx.Query("email")
	if email == "" {
		server_response.Response(ctx, http.StatusBadRequest, "pass in a valid email to recieve the otp", false, nil)
		return
	}
	otp, err := authentication.GenerateOTP(6)
	if err != nil {
		server_response.Response(ctx, http.StatusInternalServerError, "Something went wrong while generating otp", false, nil)
		return
	}
	authentication.SaveOTP(ctx, email, otp, 5*time.Minute)
	emitter.Emitter.Emit(emitter.Events.AUTH_EVENTS.RESEND_OTP, map[string]interface{}{
		"email": email,
		"otp":   strings.Split(otp, ""),
	})
	server_response.Response(ctx, http.StatusCreated, "otp sent", true, nil)
}

func HasChannelAccess(ctx *gin.Context, channel_id string, accessToCheck []channel.ChannelAdminAccess) (bool, error) {
	channelMemberRepo := workspaceRepository.GetChannelMemberRepo()
	channelMembership, err := channelMemberRepo.FindOneByFilter(map[string]interface{}{
		"userId":      utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
	})
	if err != nil || channelMembership == nil {
		err = errors.New("you are not a member of this channel")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
		return false, err
	}
	if !channelMembership.Admin {
		err = errors.New("only admins can perform this action")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return false, err
	}
	has_access := channelMembership.HasAccess(channelMembership.AdminAccess, accessToCheck)
	if !has_access {
		err = errors.New("you do not have delete access to this channel")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return false, err
	}
	return true, nil
}

func VerifyPhone(ctx *gin.Context) {
	var payload struct {
		Phone string
		Otp   string
	}
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a phone number and otp code"), StatusCode: http.StatusBadRequest})
		return
	}
	valid, err := authentication.VerifyOTP(ctx, fmt.Sprintf("%s-otp", payload.Phone), payload.Otp)
	if err != nil {
		return
	}
	if !valid {
		server_response.Response(ctx, http.StatusUnauthorized, "wrong otp provided", false, nil)
		return
	}
	success := userUseCases.UpdateUserUseCase(ctx, models.UpdateUser{Phone: &payload.Phone})
	if !success {
		return
	}
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-otp", payload.Phone))
	server_response.Response(ctx, http.StatusOK, "phone number updated", success, nil)
}

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
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/cryptography"
	"mize.app/emails"
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
	emails.SendEmail(payload.Email, "Activate your Mize account", "otp", map[string]string{"OTP": string(otp)})
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
	response, err := userRepoInstance.CreateOne(data)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("something went wrong while creating user"), StatusCode: http.StatusNotFound})
		return
	}
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-user", payload.Email))
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-otp", payload.Email))
	err = authentication.GenerateRefreshToken(ctx, *response, payload.Email, data.UserName)
	if err != nil {
		return
	}
	err = authentication.GenerateAccessToken(ctx, *response, payload.Email, data.UserName, nil)
	if err != nil {
		return
	}

	emails.SendEmail(payload.Email, "Welcome to Mize", "welcome", map[string]string{})
	server_response.Response(ctx, http.StatusCreated, "account verified", true, response)
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
	err := authentication.GenerateRefreshToken(ctx, profile.Id.Hex(), profile.Email, profile.UserName)
	if err != nil {
		return
	}
	err = authentication.GenerateAccessToken(ctx, profile.Id.Hex(), profile.Email, profile.UserName, nil)
	if err != nil {
		return
	}
	profile.Password = ""
	server_response.Response(ctx, http.StatusCreated, "login successful", true, profile)
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
			refresh_token_claims["Email"].(string), refresh_token_claims["Username"].(string), nil)
		server_response.Response(ctx, http.StatusCreated, "token generated", true, nil)
		return
	}
	err := authentication.GenerateAccessToken(ctx, refresh_token_claims["UserId"].(string),
		refresh_token_claims["Email"].(string), refresh_token_claims["Username"].(string), &workspace)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "token generated", true, nil)
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
	emails.SendEmail(email, "Activate your Mize account", "otp", map[string]string{"OTP": string(otp)})
	server_response.Response(ctx, http.StatusCreated, "otp sent", true, nil)
}

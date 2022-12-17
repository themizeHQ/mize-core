package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mize.app/app/auth/types"
	authUseCases "mize.app/app/auth/usecases"
	"mize.app/app/user/models"

	userRepo "mize.app/app/user/repository"
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/network"
	"mize.app/utils"

	"mize.app/calls/azure"
	"mize.app/emitter"
	"mize.app/realtime"
	"mize.app/repository/database/redis"
	"mize.app/server_response"
)

func CreateUser(ctx *gin.Context) {
	var payload models.User
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusBadRequest})
		return
	}
	err := authUseCases.CacheUserUseCase(ctx, payload)
	if err != nil {
		return
	}
	otp, err := authentication.GenerateOTP(6)
	if err != nil {
		server_response.Response(ctx, http.StatusInternalServerError, "something went wrong while creating user", false, nil)
		return
	}
	authentication.SaveOTP(ctx, payload.Email, otp, 5*time.Minute)
	emitter.Emitter.Emit(emitter.Events.AUTH_EVENTS.USER_CREATED, map[string]interface{}{"email": payload.Email,
		"otp": strings.Split(otp, ""), "header": "Verify Your Email Address"})
	server_response.Response(ctx, http.StatusCreated, "user created successfuly", true, nil)
}

func VerifyAccountUseCase(ctx *gin.Context) {
	var payload types.VerifyData
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
	accessToken, refreshToken, user := authUseCases.CreateUserUseCase(ctx, data)
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-user", payload.Email))
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-otp", payload.Email))
	server_response.Response(ctx, http.StatusCreated, "account verified", true, map[string]interface{}{
		"user": user,
		"tokens": map[string]string{
			"refreshToken": *refreshToken,
			"accessToken":  *accessToken,
		},
	})
}

func LoginUser(ctx *gin.Context) {
	var payload types.LoginDetails
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a json"), StatusCode: http.StatusBadRequest})
		return
	}
	refreshToken, accessToken, profile := authUseCases.LoginUserUseCase(ctx, payload)
	server_response.Response(ctx, http.StatusCreated, "login successful", true, map[string]interface{}{
		"user": profile,
		"tokens": map[string]string{
			"refreshToken": *refreshToken,
			"accessToken":  *accessToken,
		},
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
	var payload types.UpdatePassword
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in both your current password and new password"), StatusCode: http.StatusBadRequest})
		return
	}
	err := authUseCases.UpdatePasswordUseCase(ctx, payload)
	if err != nil {
		return
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

	tokenRevoked := redis.RedisRepo.FindOne(ctx, refresh_token_claims["UserId"].(string)+refresh_token)

	if tokenRevoked != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("this token has been revoked"), StatusCode: http.StatusUnauthorized})
		return
	}

	if refresh_token_claims["Type"].(string) != string(authentication.REFRESH_TOKEN) {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("invalid refresh token used"), StatusCode: http.StatusUnauthorized})
		return
	}
	invokedAllAt := redis.RedisRepo.FindOne(ctx, refresh_token_claims["UserId"].(string)+"INVALID_AT")

	if invokedAllAt != nil {
		invokedAllAtInt, err := strconv.Atoi(*invokedAllAt)
		if err != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("invalid refresh token used"), StatusCode: http.StatusUnauthorized})
			return
		}
		if int(invokedAllAtInt)-int(refresh_token_claims["iat"].(float64)) > 0 {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("revoked refresh token used"), StatusCode: http.StatusUnauthorized})
			return
		}
	}
	workspace := ctx.Query("workspace_id")
	if workspace == "" {
		accessToken, err := authentication.GenerateAccessToken(ctx, refresh_token_claims["UserId"].(string),
			refresh_token_claims["Email"].(string), refresh_token_claims["Username"].(string), refresh_token_claims["Firstname"].(string), refresh_token_claims["Lastname"].(string), nil, refresh_token_claims["ACSUserId"].(string))
		if err != nil {
			return
		}
		server_response.Response(ctx, http.StatusCreated, "token generated", true, map[string]string{
			"accessToken": accessToken,
		})
		return
	}
	accessToken, err := authentication.GenerateAccessToken(ctx, refresh_token_claims["UserId"].(string),
		refresh_token_claims["Email"].(string), refresh_token_claims["Username"].(string), refresh_token_claims["Firstname"].(string), refresh_token_claims["Lastname"].(string), &workspace, refresh_token_claims["ACSUserId"].(string))
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "token generated", true, map[string]string{
		"accessToken": accessToken,
	})
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
	err := authUseCases.ResendOTPEmailUseCase(ctx, email)
	if err != nil {
		return
	}
	server_response.Response(ctx, http.StatusCreated, "otp sent", true, nil)
}

func VerifyPhone(ctx *gin.Context) {
	var payload types.VerifyPhoneData
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("pass in a phone number and otp code"), StatusCode: http.StatusBadRequest})
		return
	}
	err := authUseCases.VerifyOTPPhoneUsecase(ctx, payload)
	if err != nil {
		return
	}
	redis.RedisRepo.DeleteOne(ctx, fmt.Sprintf("%s-otp", payload.Phone))
	server_response.Response(ctx, http.StatusOK, "phone number updated", true, nil)
}

func SignOut(ctx *gin.Context) {
	allDevices := ctx.Query("all")
	refreshToken := ctx.GetHeader("RefreshToken")
	accessToken := ctx.GetHeader("AccessToken")
	authUseCases.SignOutUseCase(ctx, allDevices, refreshToken, accessToken)
	server_response.Response(ctx, http.StatusOK, "signout success", true, nil)
}

func GoogleLogin(ctx *gin.Context) {
	state := utils.GenerateUUID()
	redis.RedisRepo.CreateEntry(ctx, ctx.ClientIP()+"GOOGLE_OAUTH_STATE", state, 20*time.Minute)
	url := authentication.SetUpConfig().AuthCodeURL(state)
	http.Redirect(ctx.Writer, ctx.Request, url, http.StatusSeeOther)
}

func GoogleCallBack(ctx *gin.Context) {
	state := redis.RedisRepo.FindOne(ctx, ctx.ClientIP()+"GOOGLE_OAUTH_STATE")
	if ctx.Query("state") != *state {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("unrecognised state used"), StatusCode: http.StatusBadRequest})
		return
	}
	code := ctx.Query("code")
	config := authentication.SetUpConfig()

	token, err := config.Exchange(ctx, code)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("code - token exchange failed"), StatusCode: http.StatusBadRequest})
		return
	}
	networkController := network.NetworkController{BaseUrl: "https://www.googleapis.com/oauth2/v2/userinfo"}
	res, err := networkController.Get("/", nil, &map[string]string{
		"access_token": token.AccessToken,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("user data fetch failed"), StatusCode: http.StatusBadRequest})
		return
	}
	userRepo := userRepo.GetUserRepo()
	var user map[string]interface{}
	json.Unmarshal([]byte(*res), &user)
	userExists, err := userRepo.FindOneByFilter(map[string]interface{}{
		"email": user["email"],
	}, options.FindOne().SetProjection(map[string]interface{}{
		"email":     1,
		"_id":       1,
		"userName":  1,
		"firstName": 1,
		"lastName":  1,
		"acsUserId": 1,
	}))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("user data fetch failed"), StatusCode: http.StatusBadRequest})
		return
	}
	var refreshToken *string
	var accessToken *string
	if userExists == nil {
		accessToken, refreshToken, userExists = authUseCases.CreateUserUseCase(ctx, models.User{
			Email:     user["email"].(string),
			UserName:  user["email"].(string),
			FirstName: user["given_name"].(string),
			LastName:  user["family_name"].(string),
			Verified:  true,
		})
		server_response.Response(ctx, http.StatusCreated, "account verified", true, map[string]interface{}{
			"user": userExists,
			"tokens": map[string]string{
				"refreshToken": *refreshToken,
				"accessToken":  *accessToken,
			},
		})
		return
	}
	rT, err := authentication.GenerateRefreshToken(ctx, userExists.Id.Hex(), userExists.Email, userExists.UserName, userExists.FirstName, userExists.LastName, userExists.ACSUserId)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not complete request"), StatusCode: http.StatusInternalServerError})
		return
	}
	aT, err := authentication.GenerateAccessToken(ctx, userExists.Id.Hex(), userExists.Email, userExists.UserName, userExists.FirstName, userExists.LastName, nil, userExists.ACSUserId)
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("could not complete request"), StatusCode: http.StatusInternalServerError})
		return
	}
	refreshToken = &rT
	accessToken = &aT
	server_response.Response(ctx, http.StatusCreated, "account verified", true, map[string]interface{}{
		"user": userExists,
		"tokens": map[string]string{
			"refreshToken": *refreshToken,
			"accessToken":  *accessToken,
		},
	})
}

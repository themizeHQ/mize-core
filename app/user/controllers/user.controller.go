package user

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	userModel "mize.app/app/user/models"
	useCases "mize.app/app/user/usecases/user"
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/emails"
	"mize.app/server_response"
	"mize.app/uuid"
)

func CacheUser(ctx *gin.Context) {
	var payload userModel.User
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return
	}
	if err := payload.Validate(); err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return
	}
	payload.RunHooks()
	response, err := useCases.CacheUserUseCase(ctx, &payload)
	if err != nil {
		return
	}
	otp, err := authentication.GenerateOTP(6)
	if err != nil {
		server_response.Response(ctx, http.StatusInternalServerError, "Something went wrong while creating user", false, nil)
		return
	}
	authentication.SaveOTP(ctx, payload.Email, otp, 5*time.Minute)
	emails.SendEmail(payload.Email, "Activate your Mize account", "otp", map[string]string{"OTP": string(otp)})
	server_response.Response(ctx, http.StatusCreated, "User created successfuly", true, response)
}

func VerifyUser(ctx *gin.Context) {
	type VerifyData struct {
		Otp   string
		Email string
	}
	var payload VerifyData
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return
	}
	valid, err := authentication.VerifyOTP(ctx, fmt.Sprintf("%s-otp", payload.Email), payload.Otp)
	if err != nil {
		return
	}
	if !valid {
		server_response.Response(ctx, http.StatusUnauthorized, "Wrong otp provided", false, nil)
		return
	}
	result, err := useCases.CreateUserUseCase(ctx, payload.Email)
	if err != nil {
		return
	}
	access_token_id := uuid.GenerateUUID()
	accessToken, err := authentication.GenerateAuthToken(ctx, authentication.ClaimsData{
		Issuer:   "mizehq",
		Type:     authentication.ACCESS_TOKEN,
		Role:     authentication.USER,
		ExpireAt: 20 * time.Minute,
		UserId:   result.InsertedID.(primitive.ObjectID).Hex(),
		TokenId:  access_token_id,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	saved_access_token := authentication.SaveAuthToken(ctx, fmt.Sprintf("%s-access-token",
		result.InsertedID.(primitive.ObjectID).Hex()), float64(time.Now().Unix()), access_token_id)
	if !saved_access_token {
		err := errors.New("failed to save access token to db")
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	refresh_token_id := uuid.GenerateUUID()
	refreshToken, err := authentication.GenerateAuthToken(ctx, authentication.ClaimsData{
		Issuer:   "mizehq",
		Type:     authentication.REFRESH_TOKEN,
		Role:     authentication.USER,
		ExpireAt: 24 * 20 * time.Hour, // 20 days
		UserId:   result.InsertedID.(primitive.ObjectID).Hex(),
		TokenId:  refresh_token_id,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	saved_refresh_token := authentication.SaveAuthToken(ctx, fmt.Sprintf("%s-refresh-token",
		result.InsertedID.(primitive.ObjectID).Hex()), float64(time.Now().Unix()), refresh_token_id)
	if !saved_refresh_token {
		err := errors.New("failed to save access token to db")
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	refresh_token_ttl, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY_TIME"))
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	access_token_ttl, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY_TIME"))
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	ctx.SetCookie(string(authentication.REFRESH_TOKEN), *refreshToken, refresh_token_ttl, "/", os.Getenv("APP_DOMAIN"), false, true)
	ctx.SetCookie(string(authentication.ACCESS_TOKEN), *accessToken, access_token_ttl, "/", os.Getenv("APP_DOMAIN"), false, true)
	emails.SendEmail(payload.Email, "Welcome to Mize", "welcome", map[string]string{})
	server_response.Response(ctx, http.StatusCreated, "Account verified", true, result)
}

func LoginUser(ctx *gin.Context) {
	type LoginDetails struct {
		Password string
		Account  string
	}
	var payload LoginDetails
	if err := ctx.ShouldBind(&payload); err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return
	}
	success, profile := useCases.LoginUserUseCase(ctx, payload.Account, payload.Password)
	if !success && profile == nil {
		server_response.Response(ctx, http.StatusNotFound, "Account not found", false, nil)
		return
	}
	if !success {
		server_response.Response(ctx, http.StatusUnauthorized, "Incorrect password", false, nil)
		return
	}
	access_token_id := uuid.GenerateUUID()
	accessToken, err := authentication.GenerateAuthToken(ctx, authentication.ClaimsData{
		Issuer:   "mizehq",
		Type:     authentication.ACCESS_TOKEN,
		Role:     authentication.USER,
		ExpireAt: 20 * time.Minute,
		UserId:   profile.Id,
		TokenId:  access_token_id,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	saved_access_token := authentication.SaveAuthToken(ctx, fmt.Sprintf("%s-access-token",
		profile.Id), float64(time.Now().Unix()), access_token_id)
	if !saved_access_token {
		err := errors.New("failed to save access token to db")
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	refresh_token_id := uuid.GenerateUUID()
	refreshToken, err := authentication.GenerateAuthToken(ctx, authentication.ClaimsData{
		Issuer:   "mizehq",
		Type:     authentication.REFRESH_TOKEN,
		Role:     authentication.USER,
		ExpireAt: 24 * 20 * time.Hour, // 20 days
		UserId:   profile.Id,
		TokenId:  refresh_token_id,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	saved_refresh_token := authentication.SaveAuthToken(ctx, fmt.Sprintf("%s-refresh-token",
		profile.Id), float64(time.Now().Unix()), refresh_token_id)
	if !saved_refresh_token {
		err := errors.New("failed to save access token to db")
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	refresh_token_ttl, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY_TIME"))
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	access_token_ttl, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY_TIME"))
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	ctx.SetCookie(string(authentication.REFRESH_TOKEN), *refreshToken, refresh_token_ttl, "/", os.Getenv("APP_DOMAIN"), false, true)
	ctx.SetCookie(string(authentication.ACCESS_TOKEN), *accessToken, access_token_ttl, "/", os.Getenv("APP_DOMAIN"), false, true)
	server_response.Response(ctx, http.StatusCreated, "Login Successful", true, profile)
}

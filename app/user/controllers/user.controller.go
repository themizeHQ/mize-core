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
	accessToken, err := authentication.GenerateAuthToken(ctx, authentication.ClaimsData{
		Issuer:   "mizehq",
		Type:     authentication.ACCESS_TOKEN,
		Role:     authentication.USER,
		ExpireAt: 20 * time.Minute,
		UserId:   result.InsertedID.(primitive.ObjectID).Hex(),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	saved_access_token := authentication.SaveAuthToken(ctx, fmt.Sprintf("%s-access-token",
		result.InsertedID.(primitive.ObjectID).Hex()), float64(time.Now().Unix()), *accessToken)
	if !saved_access_token {
		err := errors.New("failed to save access token to db")
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	refreshToken, err := authentication.GenerateAuthToken(ctx, authentication.ClaimsData{
		Issuer:   "mizehq",
		Type:     authentication.REFRESH_TOKEN,
		Role:     authentication.USER,
		ExpireAt: 24 * 20 * time.Hour, // 20 days
		UserId:   result.InsertedID.(primitive.ObjectID).Hex(),
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	saved_refresh_token := authentication.SaveAuthToken(ctx, fmt.Sprintf("%s-refresh-token",
		result.InsertedID.(primitive.ObjectID).Hex()), float64(time.Now().Unix()), *refreshToken)
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

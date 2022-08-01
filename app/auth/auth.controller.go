package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"

	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/emails"
	"mize.app/server_response"
)

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

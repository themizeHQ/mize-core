package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"

	"mize.app/app_errors"
	"mize.app/authentication"
	redis "mize.app/repository/database/redis"
	"mize.app/server_response"
	"mize.app/uuid"
)

func GenerateAccessToken(ctx *gin.Context) {
	refresh_token, err := ctx.Request.Cookie(string(authentication.REFRESH_TOKEN))
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusUnauthorized)
		return
	}

	valid_refresh_token := authentication.DecodeAuthToken(ctx, refresh_token.Value)
	if !valid_refresh_token.Valid {
		err := errors.New("invalid refresh token used")
		app_errors.ErrorHandler(ctx, err, http.StatusUnauthorized)
		return
	}

	refresh_token_claims := valid_refresh_token.Claims.(jwt.MapClaims)
	refresh_tokens := redis.RedisRepo.FindSet(ctx, fmt.Sprintf("%s-refresh-token", refresh_token_claims["UserId"]))
	var token_exists bool
	for _, val := range *refresh_tokens {
		if val == refresh_token_claims["TokenId"] {
			token_exists = true
		}
	}
	if !token_exists {
		err := errors.New("invalid refresh token used")
		app_errors.ErrorHandler(ctx, err, http.StatusUnauthorized)
		return
	}

	access_token_id := uuid.GenerateUUID()
	accessToken, err := authentication.GenerateAuthToken(ctx, authentication.ClaimsData{
		Issuer:   "mizehq",
		Type:     authentication.ACCESS_TOKEN,
		Role:     authentication.USER,
		ExpireAt: 20 * time.Minute,
		UserId:   refresh_token_claims["UserId"].(string),
		TokenId:  access_token_id,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	saved_access_token := authentication.SaveAuthToken(ctx, fmt.Sprintf("%s-access-token",
		refresh_token_claims["UserId"].(string)), float64(time.Now().Unix()), access_token_id)
	if !saved_access_token {
		err := errors.New("failed to save access token to db")
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	access_token_ttl, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY_TIME"))
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusInternalServerError)
		return
	}
	ctx.SetCookie(string(authentication.ACCESS_TOKEN), *accessToken, access_token_ttl, "/", os.Getenv("APP_DOMAIN"), false, true)
	server_response.Response(ctx, http.StatusCreated, "token generated", true, nil)
}

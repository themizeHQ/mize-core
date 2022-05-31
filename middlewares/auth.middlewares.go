package middlewares

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt"

	"mize.app/app_errors"
	"mize.app/authentication"
	redis "mize.app/repository/database/redis"
)

func AuthenticationMiddleware(ctx *gin.Context) {
	access_token, err := ctx.Request.Cookie(string(authentication.ACCESS_TOKEN))
	if err != nil {
		app_errors.ErrorHandler(ctx, err, http.StatusUnauthorized)
		return
	}

	valid_access_token := authentication.DecodeAuthToken(ctx, access_token.Value)
	if !valid_access_token.Valid {
		err := errors.New("invalid access token used 1")
		app_errors.ErrorHandler(ctx, err, http.StatusUnauthorized)
		return
	}

	access_token_claims := valid_access_token.Claims.(jwt.MapClaims)
	access_tokens := redis.RedisRepo.FindSet(ctx, fmt.Sprintf("%s-access-token", access_token_claims["UserId"]))
	fmt.Println(access_tokens)
	var token_exists bool
	for _, val := range *access_tokens {
		if val == access_token_claims["TokenId"] {
			token_exists = true
		}
	}
	if !token_exists {
		err := errors.New("invalid access token used 2")
		app_errors.ErrorHandler(ctx, err, http.StatusUnauthorized)
		return
	}
	ctx.Next()
}

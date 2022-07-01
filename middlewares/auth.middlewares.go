package middlewares

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt"

	"mize.app/app_errors"
	"mize.app/authentication"
)

func AuthenticationMiddleware(ctx *gin.Context) {
	access_token, err := ctx.Request.Cookie(string(authentication.ACCESS_TOKEN))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("no auth tokens provided"), StatusCode: http.StatusUnauthorized})
		return
	}

	valid_access_token := authentication.DecodeAuthToken(ctx, access_token.Value)
	if !valid_access_token.Valid {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("invalid access token used"), StatusCode: http.StatusUnauthorized})
		return
	}

	access_token_claims := valid_access_token.Claims.(jwt.MapClaims)
	ctx.Set("UserId", access_token_claims["UserId"])
	ctx.Set("Role", access_token_claims["Role"])
	ctx.Set("Email", access_token_claims["Email"])
	ctx.Next()
}

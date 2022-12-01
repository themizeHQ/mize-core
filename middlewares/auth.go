package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt"

	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/repository/database/redis"
)

func AuthenticationMiddleware(has_workspace bool, admin_route bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		access_token := ctx.GetHeader("Authorization")
		if access_token == "" {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("no auth tokens provided"), StatusCode: http.StatusUnauthorized})
			return
		}
		access_token = strings.Split(access_token, " ")[1]
		valid_access_token := authentication.DecodeAuthToken(ctx, access_token)
		if !valid_access_token.Valid {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("invalid access token used"), StatusCode: http.StatusUnauthorized})
			return
		}
		access_token_claims := valid_access_token.Claims.(jwt.MapClaims)

		if has_workspace {
			if access_token_claims["Workspace"] == nil {
				app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("this route is for tokens given workspace access"), StatusCode: http.StatusUnauthorized})
				return
			}
		}
		if access_token_claims["Type"] != "access_token" || access_token_claims["Issuer"] != os.Getenv("JWT_ISSUER") {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("this is not an authorized access token"), StatusCode: http.StatusUnauthorized})
			return
		}
		if admin_route {
			if access_token_claims["Role"] != string(authentication.ADMIN) {
				app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you are not authorized here"), StatusCode: http.StatusUnauthorized})
				return
			}
		}
		accessRevoked := redis.RedisRepo.FindOne(ctx, access_token_claims["UserId"].(string)+access_token)
		if accessRevoked != nil {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("this token has been revoked"), StatusCode: http.StatusUnauthorized})
			return
		}
		if access_token_claims["Type"].(string) != string(authentication.ACCESS_TOKEN) {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("invalid access token used"), StatusCode: http.StatusUnauthorized})
			return
		}
		invokedAllAt := redis.RedisRepo.FindOne(ctx, access_token_claims["UserId"].(string)+"INVALID_AT")
		if invokedAllAt != nil {
			invokedAllAtInt, err := strconv.Atoi(*invokedAllAt)
			if err != nil {
				app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("invalid access token used"), StatusCode: http.StatusUnauthorized})
				return
			}
			fmt.Println(int(invokedAllAtInt) - int(access_token_claims["exp"].(float64)))
			if int(invokedAllAtInt)-int(access_token_claims["exp"].(float64)) > 0 {
				app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("revoked access token used"), StatusCode: http.StatusUnauthorized})
				return
			}
		}
		ctx.Set("UserId", access_token_claims["UserId"])
		ctx.Set("Role", access_token_claims["Role"])
		ctx.Set("Email", access_token_claims["Email"])
		ctx.Set("Username", access_token_claims["Username"])
		ctx.Set("Workspace", access_token_claims["Workspace"])
		ctx.Set("ACSUserId", access_token_claims["ACSUserId"])
		ctx.Set("Firstname", access_token_claims["Firstname"])
		ctx.Set("Lastname", access_token_claims["Lastname"])
		ctx.Set("WorkspaceName", access_token_claims["WorkspaceName"])
		ctx.Next()
	}

}

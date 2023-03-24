package middlewares

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/golang-jwt/jwt"

	"mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/authentication"
	"mize.app/logger"
	"mize.app/repository/database/redis"
	"mize.app/utils"
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
			workspaceMemberRepo := repository.GetWorkspaceMember()
			member, err := workspaceMemberRepo.FindOneByFilter(map[string]interface{}{
				"workspaceId": *utils.HexToMongoId(ctx, access_token_claims["Workspace"].(string)),
				"userId":      *utils.HexToMongoId(ctx, access_token_claims["UserId"].(string)),
			}, options.FindOne().SetProjection(map[string]interface{}{
				"admin":       1,
				"deactivated": 1,
			}))
			if err != nil {
				e := errors.New("could not complete user authentication")
				logger.Error(e, zap.Error(err))
				app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: e, StatusCode: http.StatusInternalServerError})
				return
			}
			if member == nil {
				app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("this account no longer exists"), StatusCode: http.StatusNotFound})
				return
			}
			if member.Deactivated {
				app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("this account has been deactivated"), StatusCode: http.StatusUnauthorized})
				return
			}
			ctx.Set("Admin", member.Admin)
			if admin_route {
				if !member.Admin {
					app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("you are not authorized here"), StatusCode: http.StatusUnauthorized})
					return
				}
			}
		}
		if access_token_claims["Type"] != "access_token" || access_token_claims["Issuer"] != os.Getenv("JWT_ISSUER") {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("this is not an authorized access token"), StatusCode: http.StatusUnauthorized})
			return
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
			if int(invokedAllAtInt)-int(access_token_claims["iat"].(float64)) > 0 {
				app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: errors.New("revoked access token used"), StatusCode: http.StatusUnauthorized})
				return
			}
		}
		ctx.Set("UserId", access_token_claims["UserId"])
		ctx.Set("Role", access_token_claims["Role"])
		ctx.Set("Email", access_token_claims["Email"])
		ctx.Set("Username", access_token_claims["Username"])
		ctx.Set("Workspace", access_token_claims["Workspace"])
		ctx.Set("Firstname", access_token_claims["Firstname"])
		ctx.Set("Lastname", access_token_claims["Lastname"])
		ctx.Set("WorkspaceName", access_token_claims["WorkspaceName"])
		ctx.Next()
	}

}

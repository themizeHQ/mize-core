package authentication

import (
	"crypto"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/mongo/options"
	workspaceRepo "mize.app/app/workspace/repository"
	"mize.app/app_errors"
	"mize.app/constants/channel"
	"mize.app/cryptography"
	"mize.app/logger"
	"mize.app/repository/database/redis"
	"mize.app/utils"
)

const otpChars = "1234567890"

func GenerateOTP(length int) (string, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer), nil
}

func hashOTP(otp string) string {
	hashedOtp := cryptography.HashString(otp, nil)
	return string(hashedOtp)
}

func SaveOTP(ctx *gin.Context, email string, otp string, ttl time.Duration) bool {
	return redis.RedisRepo.CreateEntry(ctx, fmt.Sprintf("%s-otp", email), hashOTP(otp), ttl)
}

func VerifyOTP(ctx *gin.Context, key string, otp string) (bool, error) {
	data := redis.RedisRepo.FindOne(ctx, key)
	if data == nil {
		err := errors.New("otp not found")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{StatusCode: http.StatusNotFound, Err: errors.New("account not found. create a account and try again")})
		return false, err
	}
	return cryptography.VerifyData(*data, otp), nil
}

type TokenType string

const (
	ACCESS_TOKEN  TokenType = "access_token"
	REFRESH_TOKEN TokenType = "refresh_token"
)

type RoleType string

const (
	USER  RoleType = "user"
	ADMIN RoleType = "admin"
)

type ClaimsData struct {
	Issuer        string
	UserId        string
	Email         string
	Username      string
	Firstname     string
	Lastname      string
	Role          RoleType
	ExpiresAt     int64
	IssuedAt      int64
	Type          TokenType
	Workspace     *string
	WorkspaceName *string
	jwt.StandardClaims
}

// tokens
func GenerateAuthToken(ctx *gin.Context, claimsData ClaimsData) (*string, error) {
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Issuer":        claimsData.Issuer,
		"UserId":        claimsData.UserId,
		"Username":      claimsData.Username,
		"Firstname":     claimsData.Firstname,
		"Lastname":      claimsData.Lastname,
		"Role":          claimsData.Role,
		"exp":           claimsData.ExpiresAt,
		"Type":          claimsData.Type,
		"Email":         claimsData.Email,
		"Workspace":     claimsData.Workspace,
		"WorkspaceName": claimsData.WorkspaceName,
		"iat":           claimsData.IssuedAt,
	}).SignedString([]byte(os.Getenv("JWT_SIGNING_KEY")))
	if err != nil {
		return nil, err
	}
	return &tokenString, nil
}

// tokens
func GenerateCentrifugoAuthToken(ctx *gin.Context, claimsData jwt.StandardClaims) (*string, error) {
	tokenString, err := jwt.NewWithClaims(&jwt.SigningMethodHMAC{
		Name: "HS256",
		Hash: crypto.SHA256,
	}, jwt.MapClaims{
		"sub": claimsData.Subject,
		"exp": claimsData.ExpiresAt,
		"iat": claimsData.IssuedAt,
	}).SignedString([]byte(os.Getenv("CENTRIFUGO_JWT_SECRET")))
	if err != nil {
		return nil, err
	}
	return &tokenString, nil
}

func DecodeAuthToken(ctx *gin.Context, tokenString string) *jwt.Token {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SIGNING_KEY")), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			app_errors.ErrorHandler(ctx, app_errors.RequestError{StatusCode: http.StatusUnauthorized, Err: errors.New("invalid json used")})
			return nil
		}
		app_errors.ErrorHandler(ctx, app_errors.RequestError{StatusCode: http.StatusBadRequest, Err: err})
		return nil
	}
	if !token.Valid {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return nil
	}
	return token
}

func SaveAuthToken(ctx *gin.Context, key string, score float64, payload string) bool {
	return redis.RedisRepo.CreateInSet(ctx, key, score, payload)
}

func GenerateRefreshToken(ctx *gin.Context, id string, email string, username string, firstName string, lastName string) (string, error) {
	refreshToken, err := GenerateAuthToken(ctx, ClaimsData{
		Issuer:    os.Getenv("JWT_ISSUER"),
		Type:      REFRESH_TOKEN,
		Username:  username,
		ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(2400)).Unix(), // 100 days
		IssuedAt:  time.Now().Unix(),
		UserId:    id,
		Email:     email,
		Firstname: firstName,
		Lastname:  lastName,
	})
	if err != nil {
		logger.Error(err)
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return "", err
	}
	return *refreshToken, nil
}

func GenerateAccessToken(ctx *gin.Context, id string, email string, username string, firstName string, lastName string, workspace_id *string) (string, error) {
	var role RoleType = USER
	var workspaceName string
	if workspace_id != nil {
		workspaceMemberRepo := workspaceRepo.GetWorkspaceMember()
		workspaceMember, err := workspaceMemberRepo.FindOneByFilter(map[string]interface{}{"userId": utils.HexToMongoId(ctx, id), "workspaceId": utils.HexToMongoId(ctx, *workspace_id)}, options.FindOne().SetProjection(map[string]interface{}{
			"admin":         1,
			"workspaceName": 1,
		}))
		if workspaceMember == nil || err != nil {
			if err != nil {
				logger.Error(err)
			}
			err := app_errors.RequestError{StatusCode: http.StatusUnauthorized, Err: errors.New("you do not belong to this workspace")}
			app_errors.ErrorHandler(ctx, err)
			return "", err
		}
		if workspaceMember.Admin {
			role = ADMIN
		}
		workspaceName = workspaceMember.WorkspaceName
	}
	accessToken, err := GenerateAuthToken(ctx, ClaimsData{
		Issuer:        os.Getenv("JWT_ISSUER"),
		Type:          ACCESS_TOKEN,
		Username:      username,
		Firstname:     firstName,
		IssuedAt:      time.Now().Unix(),
		Lastname:      lastName,
		Role:          role,
		ExpiresAt:     time.Now().Local().Add(time.Hour * time.Duration(2)).Unix(), // 10 mins
		UserId:        id,
		Email:         email,
		Workspace:     workspace_id,
		WorkspaceName: &workspaceName,
	})
	if err != nil {
		logger.Error(err)
		err := app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized}
		app_errors.ErrorHandler(ctx, err)
		return "", err
	}
	return *accessToken, nil
}

func HasChannelAccess(ctx *gin.Context, channel_id string, accessToCheck []channel.ChannelAdminAccess) (bool, error) {
	channelMemberRepo := workspaceRepo.GetChannelMemberRepo()
	channelMembership, err := channelMemberRepo.FindOneByFilter(map[string]interface{}{
		"userId":      utils.HexToMongoId(ctx, ctx.GetString("UserId")),
		"workspaceId": utils.HexToMongoId(ctx, ctx.GetString("Workspace")),
		"channelId":   utils.HexToMongoId(ctx, channel_id),
	})
	if err != nil || channelMembership == nil {
		err = errors.New("you are not a member of this channel")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusNotFound})
		return false, err
	}
	if !channelMembership.Admin {
		err = errors.New("only admins can perform this action")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return false, err
	}
	has_access := channelMembership.HasAccess(accessToCheck)
	if !has_access {
		err = errors.New("you do not have delete access to this channel")
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return false, err
	}
	return true, nil
}

package authentication

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"mize.app/app_errors"
	"mize.app/cryptography"
	redis "mize.app/repository/database/redis"
	"mize.app/uuid"
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
		err := errors.New("account not found. create a account and try again")
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
	Issuer   string
	UserId   string
	Email    string
	Username string
	TokenId  string
	Role     RoleType
	ExpireAt time.Duration
	Type     TokenType
	jwt.StandardClaims
}

// tokens
func GenerateAuthToken(ctx *gin.Context, claimsData ClaimsData) (*string, error) {
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Issuer":   claimsData.Issuer,
		"UserId":   claimsData.UserId,
		"Username": claimsData.Username,
		"TokenId":  claimsData.TokenId,
		"Role":     claimsData.Role,
		"ExpireAt": claimsData.ExpireAt,
		"Type":     claimsData.Type,
		"Email":    claimsData.Email,
	}).SignedString([]byte(os.Getenv("JWT_SIGNING_KEY")))
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

func GenerateRefreshToken(ctx *gin.Context, id string, email string, username string) {
	refresh_token_id := uuid.GenerateUUID()
	refreshToken, err := GenerateAuthToken(ctx, ClaimsData{
		Issuer:   os.Getenv("JWT_ISSUER"),
		Type:     REFRESH_TOKEN,
		Username: username,
		Role:     USER,
		ExpireAt: 24 * 200 * time.Hour, // 200 days
		UserId:   id,
		TokenId:  refresh_token_id,
		Email:    email,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return
	}
	refresh_token_ttl, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY_TIME"))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusInternalServerError})
		return
	}
	ctx.SetCookie(string(REFRESH_TOKEN), *refreshToken, refresh_token_ttl, "/", os.Getenv("APP_DOMAIN"), false, false)
}

func GenerateAccessToken(ctx *gin.Context, id string, email string, username string) {
	access_token_id := uuid.GenerateUUID()
	accessToken, err := GenerateAuthToken(ctx, ClaimsData{
		Issuer:   os.Getenv("JWT_ISSUER"),
		Type:     ACCESS_TOKEN,
		Username: username,
		Role:     USER,
		ExpireAt: 30 * time.Minute, // 30 mins
		UserId:   id,
		TokenId:  access_token_id,
		Email:    email,
	})
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return
	}
	access_token_ttl, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY_TIME"))
	if err != nil {
		app_errors.ErrorHandler(ctx, app_errors.RequestError{Err: err, StatusCode: http.StatusUnauthorized})
		return
	}
	ctx.SetCookie(string(ACCESS_TOKEN), *accessToken, access_token_ttl, "/", os.Getenv("APP_DOMAIN"), false, true)
}

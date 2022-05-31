package authentication

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"mize.app/app_errors"
	"mize.app/cryptography"
	redis "mize.app/repository/database/redis"
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
		app_errors.ErrorHandler(ctx, app_errors.RequestError{StatusCode: http.StatusNotFound, Err: errors.New("account not found. create a account and try again")}, http.StatusNotFound)
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
	USER RoleType = "user"
)

type ClaimsData struct {
	Issuer   string
	UserId   string
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
		"Role":     claimsData.Role,
		"ExpireAt": claimsData.ExpireAt,
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
			app_errors.ErrorHandler(ctx, err, http.StatusUnauthorized)
			return nil
		}
		app_errors.ErrorHandler(ctx, err, http.StatusBadRequest)
		return nil
	}
	if !token.Valid {
		app_errors.ErrorHandler(ctx, err, http.StatusUnauthorized)
		return nil
	}
	return token
}

func SaveAuthToken(ctx *gin.Context, key string, score float64, payload string) bool {
	return redis.RedisRepo.CreateInSet(ctx, key, score, payload)
}

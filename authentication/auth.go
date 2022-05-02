package authentication

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
		app_errors.ErrorHandler(ctx, app_errors.RequestError{StatusCode: http.StatusNotFound, Err: errors.New("account not found. create a account and try again")})
		return false, err
	}
	return cryptography.VerifyData(*data, otp), nil
}

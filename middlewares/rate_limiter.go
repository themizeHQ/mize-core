package middlewares

import (
	"net/http"
	"time"

	"github.com/axiaoxin-com/ratelimiter"
	"github.com/gin-gonic/gin"
	"mize.app/db/redis"
	"mize.app/server_response"
)

func RateLimiter() gin.HandlerFunc {
	// Put a token into the token bucket every 1s
	// Maximum 1 request allowed per second
	return ratelimiter.GinRedisRatelimiter(redis.Client, ratelimiter.GinRatelimiterConfig{
		// config: how to generate a limit key
		LimitKey: func(c *gin.Context) string {
			return c.ClientIP()
		},
		// config: how to respond when limiting
		LimitedHandler: func(ctx *gin.Context) {
			server_response.Response(ctx, http.StatusTooManyRequests, "too many requests", false, nil)
			ctx.Abort()
		},
		// config: return ratelimiter token fill interval and bucket size
		TokenBucketConfig: func(*gin.Context) (time.Duration, int) {
			return time.Second * 60, 20
		},
	})
}

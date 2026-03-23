package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	ctxKeyUserID = "userID"
	ctxKeyRole   = "role"
)

// RequireAuth validates the Bearer JWT and sets userID + role in context.
func RequireAuth(cfg *core.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "authorization header missing"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := core.ParseAccessToken(tokenStr, cfg.SecretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid or expired token"})
			return
		}
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid token subject"})
			return
		}
		c.Set(ctxKeyUserID, userID)
		c.Set(ctxKeyRole, claims.Role)
		c.Next()
	}
}

// RequireSuperadmin must be used after RequireAuth.
func RequireSuperadmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ctxKeyRole)
		if role != "superadmin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "message": "superadmin access required"})
			return
		}
		c.Next()
	}
}

// RateLimit is a Redis sliding-window rate limiter keyed by "<prefix>:<client-ip>".
func RateLimit(prefix string, limit int, window time.Duration, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("%s:%s", prefix, c.ClientIP())
		ctx := context.Background()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next() // fail open on Redis error
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}
		if count > int64(limit) {
			ttl, _ := rdb.TTL(ctx, key).Result()
			c.Header("Retry-After", fmt.Sprintf("%.0f", ttl.Seconds()))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "too many requests, please slow down",
			})
			return
		}
		c.Next()
	}
}

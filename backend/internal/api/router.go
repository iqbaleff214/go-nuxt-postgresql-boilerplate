package api

import (
	"net/http"
	"time"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/api/handler"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/api/middleware"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(cfg *core.Config, db *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", cfg.FrontendURL)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(cfg, db, rdb)
	twofaSvc := service.NewTwoFAService(cfg, db, rdb, authSvc)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authH := handler.NewAuthHandler(authSvc, cfg)
	twofaH := handler.NewTwoFAHandler(twofaSvc, cfg)

	// ── Middleware helpers ─────────────────────────────────────────────────────
	authMW := middleware.RequireAuth(cfg)
	rl := func(prefix string, limit int) gin.HandlerFunc {
		return middleware.RateLimit(prefix, limit, time.Minute, rdb)
	}

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", rl("rl:register", 10), authH.Register)
			auth.POST("/verify-email", authH.VerifyEmail)
			auth.POST("/resend-verification", rl("rl:resend", 5), authH.ResendVerification)
			auth.POST("/login", rl("rl:login", 10), authH.Login)
			auth.POST("/logout", authH.Logout)
			auth.POST("/refresh", authH.Refresh)
			auth.POST("/forgot-password", rl("rl:forgot", 5), authH.ForgotPassword)
			auth.POST("/reset-password", authH.ResetPassword)
			auth.POST("/change-password", authMW, authH.ChangePassword)

			// 2FA
			auth.POST("/2fa/setup", authMW, twofaH.Setup)
			auth.POST("/2fa/confirm", authMW, twofaH.Confirm)
			auth.POST("/2fa/disable", authMW, twofaH.Disable)
			auth.POST("/2fa/verify", rl("rl:mfa", 5), twofaH.Verify)
			auth.POST("/2fa/recovery-codes/regenerate", authMW, twofaH.RegenerateCodes)
		}
	}

	return r
}

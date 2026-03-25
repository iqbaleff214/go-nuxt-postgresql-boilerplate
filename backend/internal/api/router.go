package api

import (
	"context"
	"net/http"
	"time"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/api/handler"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/api/middleware"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/service"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/templates"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// Static file serving (local storage)
	r.Static("/files", cfg.StoragePath)

	// Health check — pings DB and Redis, returns 503 if either is down
	r.GET("/health", handler.HealthHandler(db, rdb))

	// Swagger UI (dev only)
	if cfg.AppEnv != "production" {
		r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// ── Services ──────────────────────────────────────────────────────────────
	storage, err := service.NewStorageService(
		context.Background(),
		cfg.StorageBackend, cfg.StoragePath, cfg.AppBaseURL,
		cfg.S3EndpointURL, cfg.S3BucketName,
		cfg.S3AccessKey, cfg.S3SecretKey,
		cfg.S3Region, cfg.S3PublicURL,
	)
	if err != nil {
		panic("storage init failed: " + err.Error())
	}

	renderer, err := templates.NewRenderer()
	if err != nil {
		panic("template renderer init failed: " + err.Error())
	}
	redisOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		panic("parse redis URL for asynq: " + err.Error())
	}
	asynqClient := asynq.NewClient(redisOpt)
	mailer := service.NewMailer(renderer, asynqClient, cfg.AppName, cfg.FrontendURL)

	// WebSocket hub — runs in background for the lifetime of the server
	hub := ws.NewHub(rdb)
	go hub.Run(context.Background())

	authSvc := service.NewAuthService(cfg, db, rdb, mailer)
	twofaSvc := service.NewTwoFAService(cfg, db, rdb, authSvc)
	userSvc := service.NewUserService(cfg, db, storage, mailer)
	notifSvc := service.NewNotificationService(db, asynqClient)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authH := handler.NewAuthHandler(authSvc, cfg)
	twofaH := handler.NewTwoFAHandler(twofaSvc, cfg)
	profileH := handler.NewProfileHandler(userSvc)
	adminH := handler.NewAdminHandler(userSvc)
	notifH := handler.NewNotificationHandler(notifSvc)
	announcementH := handler.NewAnnouncementHandler(asynqClient)

	// ── Middleware helpers ─────────────────────────────────────────────────────
	authMW := middleware.RequireAuth(cfg)
	adminMW := middleware.RequireSuperadmin()
	rl := func(prefix string, limit int) gin.HandlerFunc {
		return middleware.RateLimit(prefix, limit, time.Minute, rdb)
	}

	// ── WebSocket ──────────────────────────────────────────────────────────────
	r.GET("/ws/notifications", func(c *gin.Context) {
		ws.ServeWS(hub, cfg, c.Writer, c.Request)
	})

	v1 := r.Group("/api/v1")
	{
		// ── Auth ─────────────────────────────────────────────────────────────
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

		// ── Profile (requires auth) ────────────────────────────────────────
		profile := v1.Group("/profile", authMW)
		{
			profile.GET("", profileH.GetProfile)
			profile.PATCH("", profileH.UpdateProfile)
			profile.POST("/avatar", profileH.UploadAvatar)
			profile.POST("/email", profileH.RequestEmailChange)
			profile.POST("/email/confirm", profileH.ConfirmEmailChange)
			profile.POST("/delete", profileH.RequestDeletion)
			profile.POST("/delete/cancel", profileH.CancelDeletion)
		}

		// ── Notifications (requires auth) ──────────────────────────────────
		notif := v1.Group("/notifications", authMW)
		{
			notif.GET("", notifH.List)
			notif.GET("/unread-count", notifH.UnreadCount)
			notif.PATCH("/read-all", notifH.MarkAllRead)
			notif.PATCH("/:id/read", notifH.MarkRead)
		}

		// ── Admin (requires superadmin) ────────────────────────────────────
		admin := v1.Group("/admin", authMW, adminMW)
		{
			users := admin.Group("/users")
			{
				users.GET("", adminH.ListUsers)
				users.POST("", adminH.CreateUser)
				users.GET("/:id", adminH.GetUser)
				users.PATCH("/:id", adminH.UpdateUser)
				users.DELETE("/:id", adminH.DeleteUser)
				users.POST("/:id/activate", adminH.ActivateUser)
				users.POST("/:id/deactivate", adminH.DeactivateUser)
				users.POST("/:id/ban", adminH.BanUser)
				users.POST("/:id/unban", adminH.UnbanUser)
			}
			admin.POST("/announcements", announcementH.Broadcast)
		}
	}

	return r
}

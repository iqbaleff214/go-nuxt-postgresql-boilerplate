package core

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// App
	AppName  string
	AppEnv   string
	Port     string
	FrontendURL string

	// Security
	SecretKey          string
	TOTPEncryptionKey  string

	// Database
	DatabaseURL string

	// JWT
	AccessTokenExpireMinutes  int
	RefreshTokenExpireDays    int
	MFAChallengeExpireMinutes int

	// Redis
	RedisURL string

	// Email
	MailProvider  string
	MailFrom      string
	MailFromName  string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
	SendGridAPIKey string
	ResendAPIKey   string

	// Storage
	StorageBackend string
	StoragePath    string
	S3EndpointURL  string
	S3BucketName   string
	S3AccessKey    string
	S3SecretKey    string
	S3Region       string
	S3PublicURL    string
}

func Load() (*Config, error) {
	if env := os.Getenv("APP_ENV"); env != "production" {
		_ = godotenv.Load()
	}

	cfg := &Config{
		AppName:     getEnv("APP_NAME", "MyApp"),
		AppEnv:      getEnv("APP_ENV", "development"),
		Port:        getEnv("PORT", "8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		SecretKey:         mustEnv("SECRET_KEY"),
		TOTPEncryptionKey: mustEnv("TOTP_ENCRYPTION_KEY"),

		DatabaseURL: mustEnv("DATABASE_URL"),

		AccessTokenExpireMinutes:  getEnvInt("ACCESS_TOKEN_EXPIRE_MINUTES", 15),
		RefreshTokenExpireDays:    getEnvInt("REFRESH_TOKEN_EXPIRE_DAYS", 7),
		MFAChallengeExpireMinutes: getEnvInt("MFA_CHALLENGE_TOKEN_EXPIRE_MINUTES", 5),

		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		MailProvider:   getEnv("MAIL_PROVIDER", "smtp"),
		MailFrom:       getEnv("MAIL_FROM", "noreply@myapp.com"),
		MailFromName:   getEnv("MAIL_FROM_NAME", "MyApp"),
		SMTPHost:       getEnv("SMTP_HOST", "localhost"),
		SMTPPort:       getEnv("SMTP_PORT", "1025"),
		SMTPUser:       getEnv("SMTP_USER", ""),
		SMTPPass:       getEnv("SMTP_PASS", ""),
		SendGridAPIKey: getEnv("SENDGRID_API_KEY", ""),
		ResendAPIKey:   getEnv("RESEND_API_KEY", ""),

		StorageBackend: getEnv("STORAGE_BACKEND", "local"),
		StoragePath:    getEnv("STORAGE_PATH", "./uploads"),
		S3EndpointURL:  getEnv("S3_ENDPOINT_URL", ""),
		S3BucketName:   getEnv("S3_BUCKET_NAME", ""),
		S3AccessKey:    getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:    getEnv("S3_SECRET_KEY", ""),
		S3Region:       getEnv("S3_REGION", ""),
		S3PublicURL:    getEnv("S3_PUBLIC_URL", ""),
	}

	if len(cfg.TOTPEncryptionKey) != 32 {
		return nil, fmt.Errorf("TOTP_ENCRYPTION_KEY must be exactly 32 characters, got %d", len(cfg.TOTPEncryptionKey))
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

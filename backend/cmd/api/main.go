// @title           Go-Nuxt Starter Kit API
// @version         1.0
// @description     REST API for the Go-Nuxt Starter Kit (auth, profile, notifications, admin).
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://github.com/404nfidv2/go-nuxt-starter-kit

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the access token.

package main

import (
	"context"
	"log"

	_ "github.com/404nfidv2/go-nuxt-starter-kit/backend/docs"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/api"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
)

func main() {
	cfg, err := core.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := core.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("database connection established")

	rdb := core.NewRedisClient(cfg.RedisURL)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	log.Println("redis connection established")

	if err := core.RunMigrations(cfg.DatabaseURL, "./migrations"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("migrations applied")

	router := api.SetupRouter(cfg, db, rdb)

	addr := ":" + cfg.Port
	log.Printf("API server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

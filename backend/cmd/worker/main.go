package main

import (
	"context"
	"log"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
)

func main() {
	cfg, err := core.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	rdb := core.NewRedisClient(cfg.RedisURL)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	log.Println("redis connection established")

	db, err := core.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("database connection established")

	// Scheduler runs cron jobs; start it in a goroutine alongside the worker
	go startScheduler(cfg.RedisURL)

	log.Println("worker starting...")
	startWorker(cfg, db, rdb)
}

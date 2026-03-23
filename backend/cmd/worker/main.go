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

	log.Println("worker starting...")
	startWorker(cfg, rdb)
}

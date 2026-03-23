package main

import (
	"log"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func startWorker(cfg *core.Config, rdb *redis.Client) {
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("invalid redis URL: %v", err)
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisOpts.Addr, Password: redisOpts.Password, DB: redisOpts.DB},
		asynq.Config{Concurrency: 10},
	)

	mux := asynq.NewServeMux()
	// Job handlers will be registered here in Phase 4 & 5

	log.Println("worker ready, processing jobs...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("asynq server error: %v", err)
	}
}

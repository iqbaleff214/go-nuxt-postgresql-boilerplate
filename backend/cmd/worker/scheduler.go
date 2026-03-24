package main

import (
	"log"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/jobs"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func startScheduler(redisURL string) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("invalid redis URL for scheduler: %v", err)
	}

	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: opts.Addr, Password: opts.Password, DB: opts.DB},
		nil,
	)

	// Daily at 02:00 UTC — purge expired tokens
	if _, err := scheduler.Register(
		"0 2 * * *",
		asynq.NewTask(jobs.TypeCleanupExpiredTokens, nil),
	); err != nil {
		log.Fatalf("register cleanup_expired_tokens: %v", err)
	}

	// Daily at 03:00 UTC — hard delete accounts soft-deleted > 30 days ago
	if _, err := scheduler.Register(
		"0 3 * * *",
		asynq.NewTask(jobs.TypeHardDeleteAccounts, nil),
	); err != nil {
		log.Fatalf("register hard_delete_accounts: %v", err)
	}

	log.Println("scheduler starting...")
	if err := scheduler.Run(); err != nil {
		log.Fatalf("asynq scheduler error: %v", err)
	}
}

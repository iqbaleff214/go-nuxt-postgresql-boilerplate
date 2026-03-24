package main

import (
	"log"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/jobs"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/service"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/ws"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func startWorker(cfg *core.Config, db *pgxpool.Pool, rdb *redis.Client) {
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("invalid redis URL: %v", err)
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisOpts.Addr, Password: redisOpts.Password, DB: redisOpts.DB},
		asynq.Config{Concurrency: 10},
	)

	emailSender := service.NewEmailSender(cfg)
	hub := ws.NewHub(rdb)

	mux := asynq.NewServeMux()
	mux.Handle(jobs.TypeSendEmail, jobs.NewEmailJobHandler(emailSender))
	mux.Handle(jobs.TypeSendNotification, jobs.NewNotificationJobHandler(hub))
	mux.Handle(jobs.TypeBroadcastAnnouncement, jobs.NewBroadcastAnnouncementHandler(db, hub))
	mux.Handle(jobs.TypeCleanupExpiredTokens, jobs.NewCleanupExpiredTokensHandler(db))
	mux.Handle(jobs.TypeHardDeleteAccounts, jobs.NewHardDeleteAccountsHandler(db))

	log.Println("worker ready, processing jobs...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("asynq server error: %v", err)
	}
}

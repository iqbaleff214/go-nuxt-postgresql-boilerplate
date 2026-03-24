package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthHandler godoc
// @Summary      Health check — pings DB and Redis
// @Tags         system
// @Produce      json
// @Success      200  {object}  object{status=string,db=string,redis=string}
// @Failure      503  {object}  object{status=string,db=string,redis=string}
// @Router       /health [get]
func HealthHandler(db *pgxpool.Pool, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := db.Ping(ctx); err != nil {
			dbStatus = "error: " + err.Error()
		}

		redisStatus := "ok"
		if err := rdb.Ping(ctx).Err(); err != nil {
			redisStatus = "error: " + err.Error()
		}

		status := http.StatusOK
		if dbStatus != "ok" || redisStatus != "ok" {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status": func() string {
				if status == http.StatusOK {
					return "ok"
				}
				return "degraded"
			}(),
			"db":    dbStatus,
			"redis": redisStatus,
		})
	}
}

package handler

import (
	"net/http"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/jobs"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

// AnnouncementHandler handles the broadcast announcement endpoint.
type AnnouncementHandler struct {
	client *asynq.Client
}

func NewAnnouncementHandler(client *asynq.Client) *AnnouncementHandler {
	return &AnnouncementHandler{client: client}
}

// POST /api/v1/admin/announcements
func (h *AnnouncementHandler) Broadcast(c *gin.Context) {
	var req struct {
		Title string `json:"title" binding:"required"`
		Body  string `json:"body"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "title is required")
		return
	}

	task, err := jobs.NewBroadcastAnnouncementTask(req.Title, req.Body)
	if err != nil {
		fail(c, http.StatusInternalServerError, "failed to create task")
		return
	}
	if _, err := h.client.EnqueueContext(c.Request.Context(), task); err != nil {
		fail(c, http.StatusInternalServerError, "failed to enqueue announcement")
		return
	}
	ok(c, http.StatusAccepted, "announcement queued", nil)
}

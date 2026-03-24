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

// Broadcast godoc
// @Summary      Broadcast a system announcement to all active users via WebSocket
// @Tags         admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  object{title=string,body=string}  true  "Announcement"
// @Success      202  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Failure      403  {object}  errorEnvelope
// @Router       /admin/announcements [post]
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

package handler

import (
	"net/http"
	"strconv"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NotificationHandler handles notification HTTP endpoints.
type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// List godoc
// @Summary      List notifications for the current user
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int  false  "Page number (default 1)"
// @Param        page_size  query  int  false  "Page size (default 20)"
// @Success      200  {object}  envelope
// @Failure      401  {object}  errorEnvelope
// @Router       /notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	userID := mustUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	items, total, err := h.svc.ListForUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// UnreadCount godoc
// @Summary      Get unread notification count
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  envelope{data=object{unread_count=int}}
// @Failure      401  {object}  errorEnvelope
// @Router       /notifications/unread-count [get]
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := mustUserID(c)
	count, err := h.svc.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, http.StatusOK, "ok", gin.H{"unread_count": count})
}

// MarkRead godoc
// @Summary      Mark a notification as read
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "Notification UUID"
// @Success      200  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Router       /notifications/{id}/read [patch]
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := mustUserID(c)
	notifID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid notification id")
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), userID, notifID); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, http.StatusOK, "notification marked as read", nil)
}

// MarkAllRead godoc
// @Summary      Mark all notifications as read
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  envelope
// @Failure      401  {object}  errorEnvelope
// @Router       /notifications/read-all [patch]
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := mustUserID(c)
	if err := h.svc.MarkAllRead(c.Request.Context(), userID); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, http.StatusOK, "all notifications marked as read", nil)
}

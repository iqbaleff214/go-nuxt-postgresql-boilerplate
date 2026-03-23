package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ctxKeyUserID = "userID"
	ctxKeyRole   = "role"
)

// mustUserID extracts the authenticated user's UUID from the Gin context.
// Panics if the middleware did not set it (programming error).
func mustUserID(c *gin.Context) uuid.UUID {
	id, _ := c.Get(ctxKeyUserID)
	return id.(uuid.UUID)
}

// mustRole extracts the authenticated user's role from the Gin context.
func mustRole(c *gin.Context) string {
	role, _ := c.Get(ctxKeyRole)
	return role.(string)
}

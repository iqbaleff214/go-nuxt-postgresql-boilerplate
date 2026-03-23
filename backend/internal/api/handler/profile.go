package handler

import (
	"net/http"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	users *service.UserService
}

func NewProfileHandler(users *service.UserService) *ProfileHandler {
	return &ProfileHandler{users: users}
}

// GET /api/v1/profile
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID := mustUserID(c)
	user, err := h.users.GetProfile(c.Request.Context(), userID)
	if err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return
	}
	ok(c, http.StatusOK, "ok", user)
}

// PATCH /api/v1/profile
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID := mustUserID(c)
	var req struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.users.UpdateProfile(c.Request.Context(), userID, req.FirstName, req.LastName, req.DisplayName, req.Bio)
	if err != nil {
		fail(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	ok(c, http.StatusOK, "Profile updated.", user)
}

// POST /api/v1/profile/avatar
func (h *ProfileHandler) UploadAvatar(c *gin.Context) {
	userID := mustUserID(c)
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		fail(c, http.StatusBadRequest, "avatar file is required")
		return
	}
	defer file.Close()

	avatarURL, err := h.users.UploadAvatar(c.Request.Context(), userID, file, header)
	if err != nil {
		fail(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	ok(c, http.StatusOK, "Avatar uploaded.", gin.H{"avatar_url": avatarURL})
}

// POST /api/v1/profile/email
func (h *ProfileHandler) RequestEmailChange(c *gin.Context) {
	userID := mustUserID(c)
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "email is required")
		return
	}
	if err := h.users.RequestEmailChange(c.Request.Context(), userID, req.Email); err != nil {
		fail(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	ok(c, http.StatusOK, "A verification link has been sent to your new email address.", nil)
}

// POST /api/v1/profile/email/confirm
func (h *ProfileHandler) ConfirmEmailChange(c *gin.Context) {
	var req struct {
		Token    string `json:"token"     binding:"required"`
		NewEmail string `json:"new_email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "token and new_email are required")
		return
	}
	if err := h.users.ConfirmEmailChange(c.Request.Context(), req.Token, req.NewEmail); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, http.StatusOK, "Email updated successfully.", nil)
}

// POST /api/v1/profile/delete
func (h *ProfileHandler) RequestDeletion(c *gin.Context) {
	userID := mustUserID(c)
	if err := h.users.RequestAccountDeletion(c.Request.Context(), userID); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, http.StatusOK, "Your account has been scheduled for deletion. Check your email for a cancellation link.", nil)
}

// POST /api/v1/profile/delete/cancel
func (h *ProfileHandler) CancelDeletion(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "token is required")
		return
	}
	if err := h.users.CancelAccountDeletion(c.Request.Context(), req.Token); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, http.StatusOK, "Account deletion cancelled. Your account has been restored.", nil)
}

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

// GetProfile godoc
// @Summary      Get current user's profile
// @Tags         profile
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  envelope
// @Failure      401  {object}  errorEnvelope
// @Failure      404  {object}  errorEnvelope
// @Router       /profile [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID := mustUserID(c)
	user, err := h.users.GetProfile(c.Request.Context(), userID)
	if err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return
	}
	ok(c, http.StatusOK, "ok", user)
}

// UpdateProfile godoc
// @Summary      Update profile fields
// @Tags         profile
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  object{first_name=string,last_name=string,display_name=string,bio=string}  false  "Fields to update"
// @Success      200  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Router       /profile [patch]
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

// UploadAvatar godoc
// @Summary      Upload avatar image (multipart/form-data, max 2 MB)
// @Tags         profile
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        avatar  formData  file  true  "Avatar image"
// @Success      200  {object}  envelope{data=object{avatar_url=string}}
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Router       /profile/avatar [post]
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

// RequestEmailChange godoc
// @Summary      Request an email address change (sends verification to new address)
// @Tags         profile
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  object{email=string}  true  "New email"
// @Success      200  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Router       /profile/email [post]
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

// ConfirmEmailChange godoc
// @Summary      Confirm email change with token from verification email
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        body  body  object{token=string,new_email=string}  true  "Token and new email"
// @Success      200  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Router       /profile/email/confirm [post]
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

// RequestDeletion godoc
// @Summary      Schedule account for deletion (30-day window, sends cancellation email)
// @Tags         profile
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  envelope
// @Failure      401  {object}  errorEnvelope
// @Failure      500  {object}  errorEnvelope
// @Router       /profile/delete [post]
func (h *ProfileHandler) RequestDeletion(c *gin.Context) {
	userID := mustUserID(c)
	if err := h.users.RequestAccountDeletion(c.Request.Context(), userID); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, http.StatusOK, "Your account has been scheduled for deletion. Check your email for a cancellation link.", nil)
}

// CancelDeletion godoc
// @Summary      Cancel scheduled account deletion using token from email
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        body  body  object{token=string}  true  "Cancellation token"
// @Success      200  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Router       /profile/delete/cancel [post]
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

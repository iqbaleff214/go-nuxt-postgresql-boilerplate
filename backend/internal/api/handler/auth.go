package handler

import (
	"net/http"
	"time"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/service"
	"github.com/gin-gonic/gin"
)

const refreshCookieName = "refresh_token"

type AuthHandler struct {
	auth *service.AuthService
	cfg  *core.Config
}

func NewAuthHandler(auth *service.AuthService, cfg *core.Config) *AuthHandler {
	return &AuthHandler{auth: auth, cfg: cfg}
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email     string `json:"email"     binding:"required"`
		Password  string `json:"password"  binding:"required"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "validation error", fieldErr("body", err.Error()))
		return
	}
	if err := h.auth.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName); err != nil {
		fail(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	ok(c, http.StatusCreated, "Registration successful. Please check your email to verify your account.", nil)
}

// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "token is required")
		return
	}
	if err := h.auth.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, http.StatusOK, "Email verified successfully.", nil)
}

// POST /api/v1/auth/resend-verification
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "email is required")
		return
	}
	_ = h.auth.ResendVerificationEmail(c.Request.Context(), req.Email)
	ok(c, http.StatusOK, "If that email is registered and unverified, a new link has been sent.", nil)
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"    binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		fail(c, http.StatusUnauthorized, err.Error())
		return
	}

	if result.Requires2FA {
		ok(c, http.StatusOK, "2FA required", gin.H{"mfa_challenge_token": result.MFAChallengeToken})
		return
	}

	h.setRefreshCookie(c, result.RefreshToken)
	ok(c, http.StatusOK, "Login successful.", gin.H{"access_token": result.AccessToken})
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	rawRefresh, _ := c.Cookie(refreshCookieName)
	_ = h.auth.Logout(c.Request.Context(), rawRefresh)
	h.clearRefreshCookie(c)
	ok(c, http.StatusOK, "Logged out.", nil)
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	rawRefresh, err := c.Cookie(refreshCookieName)
	if err != nil || rawRefresh == "" {
		fail(c, http.StatusUnauthorized, "refresh token missing")
		return
	}

	accessToken, newRefresh, err := h.auth.RefreshToken(c.Request.Context(), rawRefresh)
	if err != nil {
		h.clearRefreshCookie(c)
		fail(c, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	h.setRefreshCookie(c, newRefresh)
	ok(c, http.StatusOK, "Token refreshed.", gin.H{"access_token": accessToken})
}

// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "email is required")
		return
	}
	_ = h.auth.ForgotPassword(c.Request.Context(), req.Email)
	ok(c, http.StatusOK, "If that email is registered, a reset link has been sent.", nil)
}

// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token"    binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "token and password are required")
		return
	}
	if err := h.auth.ResetPassword(c.Request.Context(), req.Token, req.Password); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, http.StatusOK, "Password reset successfully.", nil)
}

// POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := mustUserID(c)
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "current_password and new_password are required")
		return
	}
	if err := h.auth.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, http.StatusOK, "Password changed successfully.", nil)
}

// ─── Cookie helpers ───────────────────────────────────────────────────────────

func (h *AuthHandler) setRefreshCookie(c *gin.Context, rawToken string) {
	maxAge := h.cfg.RefreshTokenExpireDays * 24 * 60 * 60
	secure := h.cfg.AppEnv == "production"
	c.SetCookie(refreshCookieName, rawToken, maxAge, "/api/v1/auth/refresh",
		"", secure, true) // HttpOnly=true
	// Also allow logout path to read the cookie
	c.SetCookie(refreshCookieName, rawToken, maxAge, "/api/v1/auth/logout",
		"", secure, true)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	expires := time.Now().Add(-time.Hour)
	_ = expires
	c.SetCookie(refreshCookieName, "", -1, "/", "", false, true)
}

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

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  object{email=string,password=string,first_name=string,last_name=string}  true  "Registration payload"
// @Success      201  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Failure      422  {object}  errorEnvelope
// @Router       /auth/register [post]
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

// VerifyEmail godoc
// @Summary      Verify email address
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  object{token=string}  true  "Verification token"
// @Success      200  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Router       /auth/verify-email [post]
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

// ResendVerification godoc
// @Summary      Resend email verification link
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  object{email=string}  true  "Email address"
// @Success      200  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Router       /auth/resend-verification [post]
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

// Login godoc
// @Summary      Login with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  object{email=string,password=string}  true  "Login credentials"
// @Success      200  {object}  envelope{data=object{access_token=string}}
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Router       /auth/login [post]
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

// Logout godoc
// @Summary      Logout (revoke refresh token)
// @Tags         auth
// @Produce      json
// @Success      200  {object}  envelope
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	rawRefresh, _ := c.Cookie(refreshCookieName)
	_ = h.auth.Logout(c.Request.Context(), rawRefresh)
	h.clearRefreshCookie(c)
	ok(c, http.StatusOK, "Logged out.", nil)
}

// Refresh godoc
// @Summary      Refresh access token using HTTP-only refresh cookie
// @Tags         auth
// @Produce      json
// @Success      200  {object}  envelope{data=object{access_token=string}}
// @Failure      401  {object}  errorEnvelope
// @Router       /auth/refresh [post]
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

// ForgotPassword godoc
// @Summary      Request a password reset link
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  object{email=string}  true  "Registered email"
// @Success      200  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Router       /auth/forgot-password [post]
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

// ResetPassword godoc
// @Summary      Reset password using token from email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  object{token=string,password=string}  true  "Reset payload"
// @Success      200  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Router       /auth/reset-password [post]
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

// ChangePassword godoc
// @Summary      Change password (requires auth)
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  object{current_password=string,new_password=string}  true  "Passwords"
// @Success      200  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Router       /auth/change-password [post]
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

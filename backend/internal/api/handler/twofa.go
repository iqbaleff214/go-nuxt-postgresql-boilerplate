package handler

import (
	"net/http"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type TwoFAHandler struct {
	twofa *service.TwoFAService
	cfg   *core.Config
}

func NewTwoFAHandler(twofa *service.TwoFAService, cfg *core.Config) *TwoFAHandler {
	return &TwoFAHandler{twofa: twofa, cfg: cfg}
}

// POST /api/v1/auth/2fa/setup
func (h *TwoFAHandler) Setup(c *gin.Context) {
	userID := mustUserID(c)
	otpauthURL, qrDataURI, err := h.twofa.Setup2FA(c.Request.Context(), userID.String())
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, http.StatusOK, "Scan the QR code with your authenticator app.", gin.H{
		"otpauth_url":  otpauthURL,
		"qr_data_uri": qrDataURI,
	})
}

// POST /api/v1/auth/2fa/confirm
func (h *TwoFAHandler) Confirm(c *gin.Context) {
	userID := mustUserID(c)
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "code is required")
		return
	}
	codes, err := h.twofa.Confirm2FA(c.Request.Context(), userID.String(), req.Code)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, http.StatusOK, "2FA enabled. Save your recovery codes — they will not be shown again.", gin.H{
		"recovery_codes": codes,
	})
}

// POST /api/v1/auth/2fa/disable
func (h *TwoFAHandler) Disable(c *gin.Context) {
	userID := mustUserID(c)
	var req struct {
		Password string `json:"password" binding:"required"`
		Code     string `json:"code"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "password and code are required")
		return
	}
	if err := h.twofa.Disable2FA(c.Request.Context(), userID.String(), req.Password, req.Code); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, http.StatusOK, "2FA disabled.", nil)
}

// POST /api/v1/auth/2fa/verify  (public — uses challenge token)
func (h *TwoFAHandler) Verify(c *gin.Context) {
	var req struct {
		MFAChallengeToken string `json:"mfa_challenge_token" binding:"required"`
		Code              string `json:"code"                binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "mfa_challenge_token and code are required")
		return
	}
	accessToken, refreshToken, err := h.twofa.VerifyMFAChallenge(c.Request.Context(), req.MFAChallengeToken, req.Code)
	if err != nil {
		fail(c, http.StatusUnauthorized, err.Error())
		return
	}

	authH := &AuthHandler{cfg: h.cfg}
	authH.setRefreshCookie(c, refreshToken)
	ok(c, http.StatusOK, "Login successful.", gin.H{"access_token": accessToken})
}

// POST /api/v1/auth/2fa/recovery-codes/regenerate
func (h *TwoFAHandler) RegenerateCodes(c *gin.Context) {
	userID := mustUserID(c)
	var req struct {
		Password string `json:"password" binding:"required"`
		Code     string `json:"code"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "password and code are required")
		return
	}
	codes, err := h.twofa.RegenerateRecoveryCodes(c.Request.Context(), userID.String(), req.Password, req.Code)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, http.StatusOK, "Recovery codes regenerated. Save these — they will not be shown again.", gin.H{
		"recovery_codes": codes,
	})
}

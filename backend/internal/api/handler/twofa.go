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

// Setup godoc
// @Summary      Begin 2FA setup — returns QR code
// @Tags         2fa
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  envelope{data=object{otpauth_url=string,qr_data_uri=string}}
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Router       /auth/2fa/setup [post]
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

// Confirm godoc
// @Summary      Confirm 2FA setup with TOTP code — returns recovery codes
// @Tags         2fa
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  object{code=string}  true  "TOTP code"
// @Success      200  {object}  envelope{data=object{recovery_codes=[]string}}
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Router       /auth/2fa/confirm [post]
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

// Disable godoc
// @Summary      Disable 2FA
// @Tags         2fa
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  object{password=string,code=string}  true  "Current password and TOTP code"
// @Success      200  {object}  envelope
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Router       /auth/2fa/disable [post]
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

// Verify godoc
// @Summary      Complete 2FA login with TOTP code or recovery code
// @Tags         2fa
// @Accept       json
// @Produce      json
// @Param        body  body  object{mfa_challenge_token=string,code=string}  true  "Challenge token and TOTP/recovery code"
// @Success      200  {object}  envelope{data=object{access_token=string}}
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Router       /auth/2fa/verify [post]
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

// RegenerateCodes godoc
// @Summary      Regenerate 2FA recovery codes
// @Tags         2fa
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  object{password=string,code=string}  true  "Current password and TOTP code"
// @Success      200  {object}  envelope{data=object{recovery_codes=[]string}}
// @Failure      400  {object}  errorEnvelope
// @Failure      401  {object}  errorEnvelope
// @Router       /auth/2fa/recovery-codes/regenerate [post]
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

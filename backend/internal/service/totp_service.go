package service

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/pquerna/otp/totp"
	gqrcode "github.com/skip2/go-qrcode"
)

type TOTPService struct {
	issuer string
}

func NewTOTPService(appName string) *TOTPService {
	return &TOTPService{issuer: appName}
}

// GenerateSecret creates a new TOTP secret for the given account email.
// Returns the raw secret, the otpauth:// URI, and a base64 PNG QR data URI.
func (s *TOTPService) GenerateSecret(accountEmail string) (secret, otpauthURL, qrDataURI string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: accountEmail,
	})
	if err != nil {
		return "", "", "", fmt.Errorf("generate totp key: %w", err)
	}

	secret = key.Secret()
	otpauthURL = key.URL()

	img, err := key.Image(200, 200)
	if err != nil {
		// Fallback: build QR from URL string
		pngBytes, qErr := gqrcode.Encode(otpauthURL, gqrcode.Medium, 200)
		if qErr == nil {
			qrDataURI = "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
		}
		return secret, otpauthURL, qrDataURI, nil
	}

	var buf bytes.Buffer
	if encErr := png.Encode(&buf, img); encErr == nil {
		qrDataURI = "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	}
	return secret, otpauthURL, qrDataURI, nil
}

// Verify returns true if code is valid for the given raw (unencrypted) secret.
func (s *TOTPService) Verify(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateRecoveryCodes returns count random 10-character alphanumeric codes,
// formatted as XXXXX-XXXXX.
func (s *TOTPService) GenerateRecoveryCodes(count int) ([]string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	codes := make([]string, count)
	for i := range codes {
		buf := make([]byte, 10)
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		result := make([]byte, 10)
		for j, b := range buf {
			result[j] = charset[int(b)%len(charset)]
		}
		codes[i] = string(result[:5]) + "-" + string(result[5:])
	}
	return codes, nil
}

package email

import "github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/templates"

type WelcomeVerifyData struct {
	templates.Base
	FirstName string
	VerifyURL string
}

type EmailVerifiedData struct {
	templates.Base
	FirstName    string
	DashboardURL string
}

type PasswordResetData struct {
	templates.Base
	FirstName string
	ResetURL  string
}

type PasswordChangedData struct {
	templates.Base
	FirstName string
	ChangedAt string
	ResetURL  string
}

type EmailChangeVerifyData struct {
	templates.Base
	FirstName  string
	NewEmail   string
	ConfirmURL string
}

type AccountDeactivatedData struct {
	templates.Base
	FirstName  string
	SupportURL string
}

type AccountBannedData struct {
	templates.Base
	FirstName  string
	Reason     string
	SupportURL string
}

type AccountDeletionConfirmData struct {
	templates.Base
	FirstName  string
	CancelURL  string
	SupportURL string
}

type NewAccountAdminData struct {
	templates.Base
	FirstName    string
	Email        string
	TempPassword string
	LoginURL     string
}

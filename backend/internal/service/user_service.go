package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/core"
	"github.com/404nfidv2/go-nuxt-starter-kit/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StorageService is the interface both local and S3 backends implement.
// Defined here to avoid a circular import; the concrete types live in Phase 7.
type StorageService interface {
	Upload(ctx context.Context, file io.Reader, path string) (publicURL string, err error)
	Delete(ctx context.Context, path string) error
	GetSignedURL(ctx context.Context, path string, expiresIn time.Duration) (string, error)
}

// UserService handles profile CRUD, avatar upload, email-change, and account deletion.
type UserService struct {
	cfg     *core.Config
	q       *repository.Queries
	tokens  *TokenService
	storage StorageService
	mailer  *Mailer
}

func NewUserService(cfg *core.Config, db *pgxpool.Pool, storage StorageService, mailer *Mailer) *UserService {
	return &UserService{
		cfg:     cfg,
		q:       repository.New(db),
		tokens:  NewTokenService(db),
		storage: storage,
		mailer:  mailer,
	}
}

// ─── Profile ─────────────────────────────────────────────────────────────────

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*repository.User, error) {
	u, err := s.q.GetUserByID(ctx, core.UUIDToPg(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &u, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, firstName, lastName, displayName, bio string) (*repository.User, error) {
	if len(displayName) > 100 {
		return nil, fmt.Errorf("display_name must be 100 characters or fewer")
	}
	if len(bio) > 500 {
		return nil, fmt.Errorf("bio must be 500 characters or fewer")
	}
	u, err := s.q.UpdateUserProfile(ctx, repository.UpdateUserProfileParams{
		ID:          core.UUIDToPg(userID),
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Bio:         core.TextToPg(&bio),
	})
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return &u, nil
}

// ─── Avatar ───────────────────────────────────────────────────────────────────

var allowedMIMEs = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

const maxAvatarSize = 2 * 1024 * 1024 // 2 MB

func (s *UserService) UploadAvatar(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > maxAvatarSize {
		return "", fmt.Errorf("avatar must be 2 MB or smaller")
	}

	// Read enough bytes to detect MIME
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])
	ext, ok := allowedMIMEs[mimeType]
	if !ok {
		return "", fmt.Errorf("unsupported file type %q; allowed: JPEG, PNG, WebP", mimeType)
	}

	// Seek back to start
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("seek file: %w", err)
	}

	// Delete old avatar if present
	user, _ := s.q.GetUserByID(ctx, core.UUIDToPg(userID))
	if user.AvatarUrl.Valid && user.AvatarUrl.String != "" {
		_ = s.storage.Delete(ctx, strings.TrimPrefix(user.AvatarUrl.String, "/files/"))
	}

	storagePath := fmt.Sprintf("avatars/%s/%s%s", userID, uuid.New(), ext)
	publicURL, err := s.storage.Upload(ctx, file, storagePath)
	if err != nil {
		return "", fmt.Errorf("upload avatar: %w", err)
	}

	if _, err := s.q.UpdateUserAvatarURL(ctx, repository.UpdateUserAvatarURLParams{
		ID:        core.UUIDToPg(userID),
		AvatarUrl: core.TextToPg(&publicURL),
	}); err != nil {
		return "", fmt.Errorf("save avatar url: %w", err)
	}
	return publicURL, nil
}

// ─── Email change ─────────────────────────────────────────────────────────────

func (s *UserService) RequestEmailChange(ctx context.Context, userID uuid.UUID, newEmail string) error {
	if !emailRe.MatchString(newEmail) {
		return fmt.Errorf("invalid email format")
	}
	// Check uniqueness
	if existing, err := s.q.GetUserByEmail(ctx, newEmail); err == nil && core.PgToUUID(existing.ID) != userID {
		return fmt.Errorf("email already in use")
	}

	user, err := s.q.GetUserByID(ctx, core.UUIDToPg(userID))
	if err != nil {
		return fmt.Errorf("user not found")
	}

	_ = s.tokens.RevokeUserTokensByType(ctx, userID, repository.TokenTypeEmailChange)
	rawToken, err := s.tokens.CreateToken(ctx, userID, repository.TokenTypeEmailChange, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("create token: %w", err)
	}
	_ = s.mailer.SendEmailChangeVerify(ctx, newEmail, user.FirstName, newEmail, rawToken)
	return nil
}

func (s *UserService) ConfirmEmailChange(ctx context.Context, rawToken, newEmail string) error {
	t, err := s.tokens.ValidateToken(ctx, rawToken, repository.TokenTypeEmailChange)
	if err != nil {
		return fmt.Errorf("invalid or expired link")
	}
	if _, err := s.q.UpdateUserEmail(ctx, repository.UpdateUserEmailParams{
		ID:    t.UserID,
		Email: newEmail,
	}); err != nil {
		return fmt.Errorf("update email: %w", err)
	}
	return s.tokens.ConsumeToken(ctx, core.PgToUUID(t.ID))
}

// ─── Account deletion ─────────────────────────────────────────────────────────

func (s *UserService) RequestAccountDeletion(ctx context.Context, userID uuid.UUID) error {
	user, err := s.q.GetUserByID(ctx, core.UUIDToPg(userID))
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if _, err := s.q.SoftDeleteUser(ctx, core.UUIDToPg(userID)); err != nil {
		return fmt.Errorf("soft delete: %w", err)
	}
	rawToken, err := s.tokens.CreateToken(ctx, userID, repository.TokenTypeDeleteCancel, 30*24*time.Hour)
	if err != nil {
		return fmt.Errorf("create cancel token: %w", err)
	}
	_ = s.mailer.SendAccountDeletionConfirm(ctx, user.Email, user.FirstName, rawToken)
	return nil
}

func (s *UserService) CancelAccountDeletion(ctx context.Context, rawToken string) error {
	t, err := s.tokens.ValidateToken(ctx, rawToken, repository.TokenTypeDeleteCancel)
	if err != nil {
		return fmt.Errorf("invalid or expired cancel link")
	}
	if _, err := s.q.CancelSoftDelete(ctx, t.UserID); err != nil {
		return fmt.Errorf("restore account: %w", err)
	}
	return s.tokens.ConsumeToken(ctx, core.PgToUUID(t.ID))
}

// ─── Superadmin: user management ─────────────────────────────────────────────

type ListUsersFilter struct {
	Role            *string
	Status          *string
	IsEmailVerified *bool
	Search          *string
}

func (s *UserService) ListUsers(ctx context.Context, f ListUsersFilter, page, pageSize int) ([]repository.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := int32((page - 1) * pageSize)

	role := ""
	if f.Role != nil {
		role = *f.Role
	}
	status := ""
	if f.Status != nil {
		status = *f.Status
	}
	verifiedStr := ""
	if f.IsEmailVerified != nil {
		if *f.IsEmailVerified {
			verifiedStr = "true"
		} else {
			verifiedStr = "false"
		}
	}
	search := ""
	if f.Search != nil {
		search = *f.Search
	}

	params := repository.ListUsersParams{
		Column1: role,
		Column2: status,
		Column3: verifiedStr,
		Column4: search,
		Limit:   int32(pageSize),
		Offset:  offset,
	}

	users, err := s.q.ListUsers(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	cParams := repository.CountUsersParams{
		Column1: role,
		Column2: status,
		Column3: verifiedStr,
		Column4: search,
	}
	total, err := s.q.CountUsers(ctx, cParams)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	return users, total, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*repository.User, error) {
	u, err := s.q.GetUserByID(ctx, core.UUIDToPg(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &u, nil
}

type AdminUpdateFields struct {
	FirstName   *string
	LastName    *string
	DisplayName *string
	Bio         *string
	Role        *string
	Status      *string
}

func (s *UserService) AdminCreateUser(ctx context.Context, email, firstName, lastName, role string) error {
	// Generate a random temp password that satisfies validation requirements
	randPart, err := core.GenerateRandomToken(8)
	if err != nil {
		return err
	}
	tempPassword := "Tmp1!" + randPart // uppercase + digit + special + random

	hashed, err := core.HashPassword(tempPassword)
	if err != nil {
		return err
	}
	user, err := s.q.CreateUser(ctx, repository.CreateUserParams{
		Email:          email,
		HashedPassword: hashed,
		FirstName:      firstName,
		LastName:       lastName,
		DisplayName:    firstName + " " + lastName,
		Role:           repository.UserRole(role),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("email already registered")
		}
		return fmt.Errorf("create user: %w", err)
	}
	// Set active + verified
	if _, err := s.q.SetEmailVerified(ctx, user.ID); err != nil {
		return fmt.Errorf("activate user: %w", err)
	}
	rawToken, err := s.tokens.CreateToken(ctx, core.PgToUUID(user.ID), repository.TokenTypePasswordReset, 72*time.Hour)
	if err != nil {
		return fmt.Errorf("create set-password token: %w", err)
	}
	_ = s.mailer.SendNewAccountAdmin(ctx, email, firstName, tempPassword, rawToken)
	return nil
}

func (s *UserService) AdminUpdateUser(ctx context.Context, callerID, targetID uuid.UUID, f AdminUpdateFields) (*repository.User, error) {
	target, err := s.q.GetUserByID(ctx, core.UUIDToPg(targetID))
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	// Guard: cannot change another superadmin's role
	if f.Role != nil && target.Role == repository.UserRoleSuperadmin && callerID != targetID {
		return nil, fmt.Errorf("cannot change role of another superadmin")
	}

	// Apply profile fields
	firstName := target.FirstName
	lastName := target.LastName
	displayName := target.DisplayName
	bio := target.Bio.String
	if f.FirstName != nil {
		firstName = *f.FirstName
	}
	if f.LastName != nil {
		lastName = *f.LastName
	}
	if f.DisplayName != nil {
		displayName = *f.DisplayName
	}
	if f.Bio != nil {
		bio = *f.Bio
	}
	updated, err := s.q.UpdateUserProfile(ctx, repository.UpdateUserProfileParams{
		ID:          core.UUIDToPg(targetID),
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Bio:         core.TextToPg(&bio),
	})
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}

	if f.Role != nil {
		updated, err = s.q.UpdateUserRole(ctx, repository.UpdateUserRoleParams{
			ID:   core.UUIDToPg(targetID),
			Role: repository.UserRole(*f.Role),
		})
		if err != nil {
			return nil, fmt.Errorf("update role: %w", err)
		}
	}
	if f.Status != nil {
		updated, err = s.q.UpdateUserStatus(ctx, repository.UpdateUserStatusParams{
			ID:     core.UUIDToPg(targetID),
			Status: repository.UserStatus(*f.Status),
		})
		if err != nil {
			return nil, fmt.Errorf("update status: %w", err)
		}
		switch repository.UserStatus(*f.Status) {
		case repository.UserStatusInactive:
			_ = s.mailer.SendAccountDeactivated(ctx, target.Email, target.FirstName)
		case repository.UserStatusBanned:
			_ = s.mailer.SendAccountBanned(ctx, target.Email, target.FirstName, "")
		}
	}
	return &updated, nil
}

func (s *UserService) AdminDeleteUser(ctx context.Context, callerID, targetID uuid.UUID) error {
	if callerID == targetID {
		return fmt.Errorf("cannot delete your own account")
	}
	target, err := s.q.GetUserByID(ctx, core.UUIDToPg(targetID))
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if target.Role == repository.UserRoleSuperadmin {
		return fmt.Errorf("cannot delete another superadmin")
	}
	return s.q.HardDeleteUser(ctx, core.UUIDToPg(targetID))
}

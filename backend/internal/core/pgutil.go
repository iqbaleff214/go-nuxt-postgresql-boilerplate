package core

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDToPg converts a google/uuid.UUID to pgtype.UUID.
func UUIDToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// PgToUUID converts a pgtype.UUID to google/uuid.UUID.
func PgToUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}

// TimeToPg converts a time.Time to pgtype.Timestamptz.
func TimeToPg(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// PgToTime converts a pgtype.Timestamptz to time.Time (returns zero if null).
func PgToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// PgTimePtr converts a pgtype.Timestamptz to *time.Time (nil if null).
func PgTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// TextToPg converts a string pointer to pgtype.Text.
func TextToPg(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// PgToTextPtr converts a pgtype.Text to *string.
func PgToTextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// NullUUID creates an invalid (null) pgtype.UUID.
func NullUUID() pgtype.UUID {
	return pgtype.UUID{}
}

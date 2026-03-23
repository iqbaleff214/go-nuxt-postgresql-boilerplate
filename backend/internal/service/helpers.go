package service

import (
	"fmt"

	"github.com/google/uuid"
)

// mustParseUUID parses a UUID string and panics if invalid.
// Use only when the string originates from a trusted source (e.g. JWT claim already validated).
func mustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("mustParseUUID: invalid UUID %q: %v", s, err))
	}
	return id
}

// Package mongoid provides ULID (Universally Unique Lexicographically Sortable Identifier)
// utilities for MongoDB document IDs.
//
// ULIDs are 128-bit identifiers that are:
//   - Lexicographically sortable (newer IDs sort after older ones)
//   - Monotonically increasing within the same millisecond
//   - URL-safe and case-insensitive
//   - Contain an embedded timestamp (first 48 bits = milliseconds since Unix epoch)
//
// This package provides:
//   - NewULID() / NewULIDWithError() for generating new ULIDs
//   - ParseULID() for parsing and validating ULID strings
//   - ULID type with Time() method to extract the embedded timestamp
//
// For ID-based CRUD operations, use the Collection methods directly:
//
//	col.FindByID(ctx, id)
//	col.UpdateByID(ctx, id, update)
//	col.DeleteByID(ctx, id)
package mongoid

import (
	"errors"
	"fmt"
	"time"

	"github.com/cloudresty/ulid"
)

// ULID represents a ULID identifier with helper methods
type ULID struct {
	id         string
	parsedULID ulid.ULID // Cached parsed ULID for efficient time extraction
}

// NewULID generates a new ULID string.
// Panics if ULID generation fails (e.g., entropy source failure).
// This is intentional: entropy failure means the random number generator is broken,
// and the application cannot safely continue generating unique IDs.
// For explicit error handling, use NewULIDWithError instead.
func NewULID() string {
	id, err := ulid.New()
	if err != nil {
		panic(fmt.Sprintf("mongoid: failed to generate ULID: %v (entropy source failure)", err))
	}
	return id
}

// NewULIDWithError generates a new ULID string and returns any error.
// Use this when you need to handle entropy source failures explicitly.
func NewULIDWithError() (string, error) {
	id, err := ulid.New()
	if err != nil {
		return "", fmt.Errorf("failed to generate ULID: %w", err)
	}
	return id, nil
}

// ParseULID parses a ULID string and returns a ULID struct.
// Returns an error if the string is not a valid ULID format.
func ParseULID(str string) (ULID, error) {
	// Validate the ULID format
	if len(str) != 26 {
		return ULID{}, errors.New("invalid ULID format: must be 26 characters")
	}

	// Parse using the underlying ulid package to validate and cache
	parsed, err := ulid.Parse(str)
	if err != nil {
		return ULID{}, fmt.Errorf("invalid ULID format: %w", err)
	}

	return ULID{id: str, parsedULID: parsed}, nil
}

// String returns the string representation of the ULID
func (u ULID) String() string {
	return u.id
}

// Time extracts the timestamp from the ULID.
// This returns the actual creation time encoded in the ULID, not the current time.
// Returns zero time if the ULID is empty or invalid.
func (u ULID) Time() time.Time {
	if u.id == "" {
		return time.Time{}
	}

	// If we have a cached parsed ULID, use it
	if u.parsedULID.GetTime() != 0 {
		return time.UnixMilli(int64(u.parsedULID.GetTime()))
	}

	// Otherwise, parse the ULID to extract the timestamp
	parsed, err := ulid.Parse(u.id)
	if err != nil {
		return time.Time{}
	}

	return time.UnixMilli(int64(parsed.GetTime()))
}

// IsZero returns true if the ULID is zero/empty
func (u ULID) IsZero() bool {
	return u.id == ""
}

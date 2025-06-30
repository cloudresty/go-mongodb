package mongoid

import (
	"testing"
)

func TestNewULID(t *testing.T) {
	id := NewULID()

	if len(id) != 26 {
		t.Errorf("Expected ULID length of 26, got %d", len(id))
	}

	if id == "" {
		t.Error("ULID should not be empty")
	}
}

func TestParseULID(t *testing.T) {
	// Test valid ULID
	validID := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	ulid, err := ParseULID(validID)
	if err != nil {
		t.Errorf("Expected no error for valid ULID, got %v", err)
	}

	if ulid.String() != validID {
		t.Errorf("Expected ULID string to be %s, got %s", validID, ulid.String())
	}

	// Test invalid ULID
	invalidID := "invalid"
	_, err = ParseULID(invalidID)
	if err == nil {
		t.Error("Expected error for invalid ULID")
	}
}

func TestULIDMethods(t *testing.T) {
	// Test Zero ULID
	zeroULID := ULID{}
	if !zeroULID.IsZero() {
		t.Error("Zero ULID should report as zero")
	}

	// Test non-zero ULID
	id := NewULID()
	ulid, _ := ParseULID(id)
	if ulid.IsZero() {
		t.Error("Non-zero ULID should not report as zero")
	}

	// Test String method
	if ulid.String() != id {
		t.Errorf("Expected String() to return %s, got %s", id, ulid.String())
	}

	// Test Time method (just ensure it doesn't panic)
	_ = ulid.Time()
}

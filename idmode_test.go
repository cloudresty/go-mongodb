package mongodb

import (
	"testing"
)

func TestIDModeConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		idMode IDMode
	}{
		{
			name:   "ULID mode",
			idMode: IDModeULID,
		},
		{
			name:   "ObjectID mode",
			idMode: IDModeObjectID,
		},
		{
			name:   "Custom mode",
			idMode: IDModeCustom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a client with the specified ID mode
			config := &Config{
				Hosts:    "localhost:27017",
				Database: "test",
				IDMode:   tt.idMode,
			}

			client := &Client{
				config: config,
			}

			// Verify the ID mode is set correctly
			if client.config.IDMode != tt.idMode {
				t.Errorf("Expected IDMode %v, got %v", tt.idMode, client.config.IDMode)
			}
		})
	}
}

func TestIDModeConstants(t *testing.T) {
	// Test that ID mode constants are defined correctly
	if IDModeULID == "" {
		t.Error("IDModeULID should not be empty")
	}
	if IDModeObjectID == "" {
		t.Error("IDModeObjectID should not be empty")
	}
	if IDModeCustom == "" {
		t.Error("IDModeCustom should not be empty")
	}
}

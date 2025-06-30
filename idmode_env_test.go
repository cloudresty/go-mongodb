package mongodb

import (
	"os"
	"testing"
)

func TestIDModeEnvironmentConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		expect IDMode
		valid  bool
	}{
		{
			name:   "ULID via env",
			envVal: "ulid",
			expect: IDModeULID,
			valid:  true,
		},
		{
			name:   "ObjectID via env",
			envVal: "objectid",
			expect: IDModeObjectID,
			valid:  true,
		},
		{
			name:   "Custom via env",
			envVal: "custom",
			expect: IDModeCustom,
			valid:  true,
		},
		{
			name:   "Invalid mode",
			envVal: "invalid",
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("MONGODB_ID_MODE", tt.envVal)
			defer os.Unsetenv("MONGODB_ID_MODE")

			// Load config from environment using functional options
			client, err := NewClient(FromEnv())

			if tt.valid {
				if err != nil {
					t.Fatalf("Expected valid config, got error: %v", err)
				}
				config := client.config
				defer client.Close()
				if config.IDMode != tt.expect {
					t.Errorf("Expected IDMode %v, got %v", tt.expect, config.IDMode)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for invalid IDMode, but got nil")
				}
			}
		})
	}
}

func TestDefaultIDMode(t *testing.T) {
	// Clear any existing environment variable
	os.Unsetenv("MONGODB_ID_MODE")

	client, err := NewClient(FromEnv())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	config := client.config
	if config.IDMode != IDModeULID {
		t.Errorf("Expected default IDMode to be ULID, got: %v", config.IDMode)
	}
}

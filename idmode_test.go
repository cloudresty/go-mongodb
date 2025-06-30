package mongodb

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestIDModeConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		idMode     IDMode
		expectType string
	}{
		{
			name:       "ULID mode",
			idMode:     IDModeULID,
			expectType: "string",
		},
		{
			name:       "ObjectID mode",
			idMode:     IDModeObjectID,
			expectType: "objectid",
		},
		{
			name:       "Custom mode",
			idMode:     IDModeCustom,
			expectType: "custom",
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

			// Test document enhancement
			doc := map[string]interface{}{
				"name":  "test",
				"value": 123,
			}

			enhanced := client.enhanceDocument(doc)

			// Check the result based on mode
			switch tt.idMode {
			case IDModeULID:
				if id, ok := enhanced["_id"].(string); !ok || len(id) != 26 {
					t.Errorf("Expected ULID string of length 26, got %T: %v", enhanced["_id"], enhanced["_id"])
				}
			case IDModeObjectID:
				if _, ok := enhanced["_id"].(bson.ObjectID); !ok {
					t.Errorf("Expected ObjectID, got %T: %v", enhanced["_id"], enhanced["_id"])
				}
			case IDModeCustom:
				if _, exists := enhanced["_id"]; exists {
					t.Errorf("Expected no _id field, but found: %v", enhanced["_id"])
				}
			}

			// Verify other fields are present
			if enhanced["name"] != "test" {
				t.Errorf("Expected name field to be preserved, got: %v", enhanced["name"])
			}
			// Check if value exists and has the right value (might be different type after marshal/unmarshal)
			if val, exists := enhanced["value"]; !exists {
				t.Errorf("Expected value field to be preserved")
			} else {
				// Value might be int32 or int64 after BSON marshal/unmarshal
				switch v := val.(type) {
				case int32:
					if v != 123 {
						t.Errorf("Expected value 123, got: %v", v)
					}
				case int64:
					if v != 123 {
						t.Errorf("Expected value 123, got: %v", v)
					}
				case int:
					if v != 123 {
						t.Errorf("Expected value 123, got: %v", v)
					}
				default:
					t.Errorf("Expected value 123, got unexpected type %T: %v", v, v)
				}
			}
			if enhanced["created_at"] == nil {
				t.Errorf("Expected created_at field to be added")
			}
			if enhanced["updated_at"] == nil {
				t.Errorf("Expected updated_at field to be added")
			}
		})
	}
}

func TestUserProvidedID(t *testing.T) {
	config := &Config{
		Hosts:    "localhost:27017",
		Database: "test",
		IDMode:   IDModeULID, // Even with ULID mode
	}

	client := &Client{
		config: config,
	}

	// Test with user-provided _id
	doc := map[string]interface{}{
		"_id":  "user-provided-id",
		"name": "test",
	}

	enhanced := client.enhanceDocument(doc)

	// User-provided ID should be preserved
	if enhanced["_id"] != "user-provided-id" {
		t.Errorf("Expected user-provided ID to be preserved, got: %v", enhanced["_id"])
	}
}

func TestIDModeValidation(t *testing.T) {
	tests := []struct {
		mode  string
		valid bool
	}{
		{"ulid", true},
		{"objectid", true},
		{"custom", true},
		{"invalid", false},
		{"", false},
		{"ULID", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			result := isValidIDMode(tt.mode)
			if result != tt.valid {
				t.Errorf("isValidIDMode(%q) = %v, want %v", tt.mode, result, tt.valid)
			}
		})
	}
}

package mongodb

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromEnv(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Save current environment variables that we'll modify
	envVars := map[string]string{
		"MONGODB_HOSTS":             "localhost:27017",
		"MONGODB_USERNAME":          "admin",
		"MONGODB_PASSWORD":          "password",
		"MONGODB_DATABASE":          "testdb",
		"MONGODB_AUTH_DATABASE":     "admin",
		"MONGODB_MAX_POOL_SIZE":     "50",
		"MONGODB_CONNECT_TIMEOUT":   "10s",
		"MONGODB_DIRECT_CONNECTION": "true", // Add direct connection to fix replica set issues
	}

	savedValues := make(map[string]string)
	for key := range envVars {
		savedValues[key] = os.Getenv(key)
	}

	// Set environment variables
	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		// Restore environment variables
		for key := range envVars {
			if value, exists := savedValues[key]; exists && value != "" {
				_ = os.Setenv(key, value)
			} else {
				_ = os.Unsetenv(key)
			}
		}
	}()

	// Test the configuration loading using the functional options pattern
	client, err := NewClient(FromEnv())
	if err != nil {
		t.Fatalf("Failed to create client with environment config: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Get the config from the client to verify values
	config := client.config

	// Verify loaded values
	if config.Hosts != "localhost:27017" {
		t.Errorf("Expected hosts 'localhost:27017', got '%s'", config.Hosts)
	}
	if config.Database != "testdb" {
		t.Errorf("Expected database 'testdb', got '%s'", config.Database)
	}
	if config.Username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", config.Username)
	}
	if config.Password != "password" {
		t.Errorf("Expected password 'password', got '%s'", config.Password)
	}
	if config.AuthDatabase != "admin" {
		t.Errorf("Expected auth database 'admin', got '%s'", config.AuthDatabase)
	}
	if config.MaxPoolSize != 50 {
		t.Errorf("Expected max pool size 50, got %d", config.MaxPoolSize)
	}
	if config.ConnectTimeout != 10*time.Second {
		t.Errorf("Expected connect timeout 10s, got %v", config.ConnectTimeout)
	}
}

func TestLoadFromEnvWithPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set test environment variables with custom prefix
	envVars := map[string]string{
		"MYAPP_MONGODB_HOSTS":             "localhost:27017",
		"MYAPP_MONGODB_USERNAME":          "admin",
		"MYAPP_MONGODB_PASSWORD":          "password",
		"MYAPP_MONGODB_DATABASE":          "prefixdb",
		"MYAPP_MONGODB_AUTH_DATABASE":     "admin",
		"MYAPP_MONGODB_DIRECT_CONNECTION": "true", // Add direct connection to fix replica set issues
	}

	savedValues := make(map[string]string)
	for key := range envVars {
		savedValues[key] = os.Getenv(key)
	}

	// Set environment variables
	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		// Restore environment variables
		for key := range envVars {
			if value, exists := savedValues[key]; exists && value != "" {
				_ = os.Setenv(key, value)
			} else {
				_ = os.Unsetenv(key)
			}
		}
	}()

	// Test the configuration loading using the functional options pattern with prefix
	client, err := NewClient(FromEnvWithPrefix("MYAPP_"))
	if err != nil {
		t.Fatalf("Failed to create client with environment config with prefix: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Get the config from the client to verify values
	config := client.config

	// Verify loaded values
	if config.Hosts != "localhost:27017" {
		t.Errorf("Expected hosts 'localhost:27017', got '%s'", config.Hosts)
	}
	if config.Database != "prefixdb" {
		t.Errorf("Expected database 'prefixdb', got '%s'", config.Database)
	}
	if config.Username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", config.Username)
	}
	if config.Password != "password" {
		t.Errorf("Expected password 'password', got '%s'", config.Password)
	}
	if config.AuthDatabase != "admin" {
		t.Errorf("Expected auth database 'admin', got '%s'", config.AuthDatabase)
	}
}

func TestBuildConnectionURI(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "basic configuration",
			config: &Config{
				Hosts:    "localhost:27017",
				Username: "",
				Password: "",
			},
			expected: "mongodb://localhost:27017",
		},
		{
			name: "with authentication",
			config: &Config{
				Hosts:        "example.com:27017",
				Username:     "user",
				Password:     "pass",
				AuthDatabase: "admin",
			},
			expected: "mongodb://user:pass@example.com:27017?authSource=admin",
		},
		{
			name: "with replica set",
			config: &Config{
				Hosts:      "cluster.example.com:27017",
				ReplicaSet: "rs0",
			},
			expected: "mongodb://cluster.example.com:27017?replicaSet=rs0",
		},
		{
			name: "full configuration",
			config: &Config{
				Hosts:        "secure.example.com:27017",
				Username:     "secure",
				Password:     "password",
				AuthDatabase: "admin",
				ReplicaSet:   "rs0",
			},
			expected: "mongodb://secure:password@secure.example.com:27017?authSource=admin&replicaSet=rs0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test URI building using the Config method
			result := tt.config.BuildConnectionURI()
			if result != tt.expected {
				t.Errorf("Expected URI '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConfigBuildConnectionURI(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "basic config",
			config: &Config{
				Hosts:    "localhost:27017",
				Database: "testdb",
			},
			expected: "mongodb://localhost:27017/testdb",
		},
		{
			name: "with credentials",
			config: &Config{
				Hosts:        "localhost:27017",
				Username:     "user",
				Password:     "pass",
				Database:     "testdb",
				AuthDatabase: "admin",
			},
			expected: "mongodb://user:pass@localhost:27017/testdb?authSource=admin",
		},
		{
			name: "with replica set",
			config: &Config{
				Hosts:      "localhost:27017",
				Database:   "testdb",
				ReplicaSet: "rs0",
			},
			expected: "mongodb://localhost:27017/testdb?replicaSet=rs0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.BuildConnectionURI()
			if result != tt.expected {
				t.Errorf("Expected URI %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEnvDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Save current environment variables
	envVarsToSave := []string{
		"MONGODB_HOSTS", "MONGODB_USERNAME", "MONGODB_PASSWORD",
		"MONGODB_DATABASE", "MONGODB_AUTH_DATABASE", "MONGODB_REPLICA_SET",
		"MONGODB_MAX_POOL_SIZE", "MONGODB_MIN_POOL_SIZE", "MONGODB_CONNECT_TIMEOUT", "MONGODB_APP_NAME",
	}

	savedValues := make(map[string]string)
	for _, envVar := range envVarsToSave {
		savedValues[envVar] = os.Getenv(envVar)
		_ = os.Unsetenv(envVar)
	}
	defer func() {
		// Restore environment variables
		for _, envVar := range envVarsToSave {
			if value, exists := savedValues[envVar]; exists && value != "" {
				_ = os.Setenv(envVar, value)
			} else {
				_ = os.Unsetenv(envVar)
			}
		}
	}()

	// Test the configuration loading using the functional options pattern
	client, err := NewClient(FromEnv())
	if err != nil {
		t.Fatalf("Failed to create client with default config: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Get the config from the client to verify defaults
	config := client.config

	// Verify defaults
	if config.Hosts != "localhost:27017" {
		t.Errorf("Expected default hosts 'localhost:27017', got '%s'", config.Hosts)
	}
	if config.Database != "app" {
		t.Errorf("Expected default database 'app', got '%s'", config.Database)
	}
	if config.AuthDatabase != "admin" {
		t.Errorf("Expected default auth database 'admin', got '%s'", config.AuthDatabase)
	}
	if config.MaxPoolSize != 100 {
		t.Errorf("Expected default max pool size 100, got %d", config.MaxPoolSize)
	}
	if config.MinPoolSize != 5 {
		t.Errorf("Expected default min pool size 5, got %d", config.MinPoolSize)
	}
	if config.ConnectTimeout != 10*time.Second {
		t.Errorf("Expected default connect timeout 10s, got %v", config.ConnectTimeout)
	}
	if config.AppName != "go-mongodb-app" {
		t.Errorf("Expected default app name 'go-mongodb-app', got '%s'", config.AppName)
	}
}

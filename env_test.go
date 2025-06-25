package mongodb

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromEnv(t *testing.T) {
	// Save current environment variables that we'll modify
	envVars := map[string]string{
		"MONGODB_HOST":            "testhost",
		"MONGODB_PORT":            "27018",
		"MONGODB_USERNAME":        "testuser",
		"MONGODB_PASSWORD":        "testpass",
		"MONGODB_DATABASE":        "testdb",
		"MONGODB_AUTH_DATABASE":   "testauth",
		"MONGODB_REPLICA_SET":     "testrs",
		"MONGODB_MAX_POOL_SIZE":   "50",
		"MONGODB_CONNECT_TIMEOUT": "10s",
	}

	savedValues := make(map[string]string)
	for key := range envVars {
		savedValues[key] = os.Getenv(key)
	}

	// Set environment variables
	for key, value := range envVars {
		os.Setenv(key, value)
	}
	defer func() {
		// Restore environment variables
		for key := range envVars {
			if value, exists := savedValues[key]; exists && value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Test the configuration loading using our LoadConfig function
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config from environment: %v", err)
	}

	// Verify loaded values
	if config.Host != "testhost" {
		t.Errorf("Expected host 'testhost', got '%s'", config.Host)
	}
	if config.Port != 27018 {
		t.Errorf("Expected port 27018, got %d", config.Port)
	}
	if config.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", config.Username)
	}
	if config.Password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", config.Password)
	}
	if config.Database != "testdb" {
		t.Errorf("Expected database 'testdb', got '%s'", config.Database)
	}
	if config.AuthDatabase != "testauth" {
		t.Errorf("Expected auth database 'testauth', got '%s'", config.AuthDatabase)
	}
	if config.ReplicaSet != "testrs" {
		t.Errorf("Expected replica set 'testrs', got '%s'", config.ReplicaSet)
	}
	if config.MaxPoolSize != 50 {
		t.Errorf("Expected max pool size 50, got %d", config.MaxPoolSize)
	}
	if config.ConnectTimeout != 10*time.Second {
		t.Errorf("Expected connect timeout 10s, got %v", config.ConnectTimeout)
	}
}

func TestLoadFromEnvWithPrefix(t *testing.T) {
	// Set test environment variables with custom prefix
	envVars := map[string]string{
		"MYAPP_MONGODB_HOST":     "prefixhost",
		"MYAPP_MONGODB_PORT":     "27019",
		"MYAPP_MONGODB_DATABASE": "prefixdb",
	}

	savedValues := make(map[string]string)
	for key := range envVars {
		savedValues[key] = os.Getenv(key)
	}

	// Set environment variables
	for key, value := range envVars {
		os.Setenv(key, value)
	}
	defer func() {
		// Restore environment variables
		for key := range envVars {
			if value, exists := savedValues[key]; exists && value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Test the configuration loading using our LoadConfigWithPrefix function
	config, err := LoadConfigWithPrefix("MYAPP_")
	if err != nil {
		t.Fatalf("Failed to load config from environment with prefix: %v", err)
	}

	// Verify loaded values
	if config.Host != "prefixhost" {
		t.Errorf("Expected host 'prefixhost', got '%s'", config.Host)
	}
	if config.Port != 27019 {
		t.Errorf("Expected port 27019, got %d", config.Port)
	}
	if config.Database != "prefixdb" {
		t.Errorf("Expected database 'prefixdb', got '%s'", config.Database)
	}

	// Verify defaults are still applied for unset variables
	if config.Username != "" {
		t.Errorf("Expected default empty username, got '%s'", config.Username)
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
				Host:     "localhost",
				Port:     27017,
				Username: "",
				Password: "",
			},
			expected: "mongodb://localhost:27017",
		},
		{
			name: "with authentication",
			config: &Config{
				Host:         "example.com",
				Port:         27017,
				Username:     "user",
				Password:     "pass",
				AuthDatabase: "admin",
			},
			expected: "mongodb://user:pass@example.com:27017?authSource=admin",
		},
		{
			name: "with replica set",
			config: &Config{
				Host:       "cluster.example.com",
				Port:       27017,
				ReplicaSet: "rs0",
			},
			expected: "mongodb://cluster.example.com:27017?replicaSet=rs0",
		},
		{
			name: "full configuration",
			config: &Config{
				Host:         "secure.example.com",
				Port:         27017,
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
			// Create client with config to test URI building
			client := &Client{config: tt.config}
			result := client.buildConnectionURI()
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
				Host:     "localhost",
				Port:     27017,
				Database: "testdb",
			},
			expected: "mongodb://localhost:27017/testdb",
		},
		{
			name: "with credentials",
			config: &Config{
				Host:         "localhost",
				Port:         27017,
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
				Host:       "localhost",
				Port:       27017,
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
	// Save current environment variables
	envVarsToSave := []string{
		"MONGODB_HOST", "MONGODB_PORT", "MONGODB_USERNAME", "MONGODB_PASSWORD",
		"MONGODB_DATABASE", "MONGODB_AUTH_DATABASE", "MONGODB_REPLICA_SET",
		"MONGODB_MAX_POOL_SIZE", "MONGODB_MIN_POOL_SIZE", "MONGODB_CONNECT_TIMEOUT", "MONGODB_APP_NAME",
	}

	savedValues := make(map[string]string)
	for _, envVar := range envVarsToSave {
		savedValues[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}
	defer func() {
		// Restore environment variables
		for _, envVar := range envVarsToSave {
			if value, exists := savedValues[envVar]; exists && value != "" {
				os.Setenv(envVar, value)
			} else {
				os.Unsetenv(envVar)
			}
		}
	}()

	// Test the configuration loading using our LoadConfig function
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config with defaults: %v", err)
	}

	// Verify defaults
	if config.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", config.Host)
	}
	if config.Port != 27017 {
		t.Errorf("Expected default port 27017, got %d", config.Port)
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

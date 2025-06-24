package mongodb

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cloudresty/go-env"
)

func TestDebugConfigLoading(t *testing.T) {
	// Print current environment variables
	t.Logf("Environment variables:")
	envVars := []string{
		"MONGODB_HOST", "MONGODB_PORT", "MONGODB_DATABASE",
		"MONGODB_USERNAME", "MONGODB_PASSWORD", "MONGODB_AUTH_DATABASE",
	}
	for _, env := range envVars {
		t.Logf("  %s = %s", env, os.Getenv(env))
	}

	// Load config and print it
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	t.Logf("Loaded config:")
	t.Logf("  Host: %s", config.Host)
	t.Logf("  Port: %d", config.Port)
	t.Logf("  Username: %s", config.Username)
	t.Logf("  Password: %s", config.Password)
	t.Logf("  Database: %s", config.Database)
	t.Logf("  AuthDatabase: %s", config.AuthDatabase)

	// Build URI and print it
	uri := config.BuildConnectionURI()
	t.Logf("Built URI: %s", uri)
}

func TestActualConnection(t *testing.T) {
	// Load config from environment
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Increase timeout for connection
	config.ConnectTimeout = 30 * time.Second
	config.ServerSelectTimeout = 30 * time.Second

	t.Logf("Attempting to connect with URI: %s", config.BuildConnectionURI())

	// Try to create client
	client, err := NewClientWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Test if connected
	if !client.IsConnected() {
		t.Error("Client is not connected")
	} else {
		t.Log("Successfully connected to MongoDB!")
	}
}

func TestTimeoutConfiguration(t *testing.T) {
	// Test timeout configuration and MongoDB driver v2 changes
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	t.Logf("Current timeout configuration:")
	t.Logf("  ConnectTimeout: %v", config.ConnectTimeout)
	t.Logf("  ServerSelectTimeout: %v", config.ServerSelectTimeout)
	t.Logf("  SocketTimeout: %v", config.SocketTimeout)

	// Test with different timeout configurations
	testCases := []struct {
		name                string
		connectTimeout      time.Duration
		serverSelectTimeout time.Duration
		socketTimeout       time.Duration
		expectSuccess       bool
	}{
		{
			name:                "very short timeouts",
			connectTimeout:      1 * time.Second,
			serverSelectTimeout: 1 * time.Second,
			socketTimeout:       1 * time.Second,
			expectSuccess:       false, // Should timeout
		},
		{
			name:                "reasonable timeouts",
			connectTimeout:      10 * time.Second,
			serverSelectTimeout: 10 * time.Second,
			socketTimeout:       10 * time.Second,
			expectSuccess:       true, // Should work
		},
		{
			name:                "generous timeouts",
			connectTimeout:      30 * time.Second,
			serverSelectTimeout: 30 * time.Second,
			socketTimeout:       30 * time.Second,
			expectSuccess:       true, // Should definitely work
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testConfig := *config // Copy config
			testConfig.ConnectTimeout = tc.connectTimeout
			testConfig.ServerSelectTimeout = tc.serverSelectTimeout
			testConfig.SocketTimeout = tc.socketTimeout

			t.Logf("Testing with timeouts: connect=%v, select=%v, socket=%v",
				tc.connectTimeout, tc.serverSelectTimeout, tc.socketTimeout)

			start := time.Now()
			client, err := NewClientWithConfig(&testConfig)
			duration := time.Since(start)

			t.Logf("Connection attempt took: %v", duration)

			if tc.expectSuccess {
				if err != nil {
					t.Logf("Expected success but got error: %v", err)
				} else {
					t.Logf("Success: Connected successfully")
					client.Close()
				}
			} else {
				if err != nil {
					t.Logf("Expected failure and got error: %v", err)
				} else {
					t.Logf("Expected failure but connection succeeded")
					client.Close()
				}
			}
		})
	}
}

func TestMongoDriverV2TimeoutBehavior(t *testing.T) {
	// Research MongoDB driver v2 timeout behavior and potential issues
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	t.Logf("=== MongoDB Driver v2 Timeout Investigation ===")

	// Test 1: Check if SetTimeout() in driver v2 works as expected
	t.Logf("1. Testing SetTimeout() behavior in MongoDB driver v2")

	testClient, err := NewClientWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}
	defer testClient.Close()

	// Test 2: Check context timeout vs driver timeout interaction
	t.Logf("2. Testing context timeout vs driver timeout interaction")

	// Create a very short context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err = testClient.Ping(ctx)
	duration := time.Since(start)

	t.Logf("Ping with 100ms context timeout took: %v, error: %v", duration, err)
	// Test 3: Check default vs custom timeout behavior
	t.Logf("3. Testing if SocketTimeout affects operations")

	// Test with short socket timeout
	configShort := *config
	configShort.SocketTimeout = 1 * time.Second

	clientShort, err := NewClientWithConfig(&configShort)
	if err != nil {
		t.Fatalf("Failed to create short timeout client: %v", err)
	}
	defer clientShort.Close()

	// Try an operation that might timeout
	collectionShort := clientShort.Collection("timeout_test")

	start = time.Now()
	_, err = collectionShort.InsertOne(context.Background(), map[string]any{"test": "data"})
	duration = time.Since(start)

	t.Logf("Insert with 1s SocketTimeout took: %v, error: %v", duration, err)

	// Test 4: Check what happens with very long operations
	t.Logf("4. Testing driver v2 timeout semantics")
	t.Logf("   Driver v2 uses SetTimeout() for operation-level timeouts")
	t.Logf("   ConnectTimeout: %v", config.ConnectTimeout)
	t.Logf("   ServerSelectTimeout: %v", config.ServerSelectTimeout)
	t.Logf("   SocketTimeout (SetTimeout): %v", config.SocketTimeout)
}

func TestMongoDriverV2SpecificIssues(t *testing.T) {
	t.Logf("=== MongoDB Driver v2 Specific Timeout Issues Investigation ===")

	// Key changes in MongoDB driver v2:
	// 1. SetTimeout() replaces multiple operation timeouts
	// 2. Better context propagation
	// 3. Simplified timeout model

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test 1: Driver v2 SetTimeout() vs operation context
	t.Logf("1. Testing SetTimeout() interaction with operation contexts")

	// Create client with very short SocketTimeout
	testConfig := *config
	testConfig.SocketTimeout = 500 * time.Millisecond

	client, err := NewClientWithConfig(&testConfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Test with different context timeouts
	timeouts := []time.Duration{
		100 * time.Millisecond, // Shorter than driver timeout
		1 * time.Second,        // Longer than driver timeout
		10 * time.Second,       // Much longer than driver timeout
	}

	for _, timeout := range timeouts {
		t.Run(fmt.Sprintf("context_%v", timeout), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			start := time.Now()
			err := client.Ping(ctx)
			duration := time.Since(start)

			t.Logf("Ping with context timeout %v took %v, error: %v", timeout, duration, err)

			// In driver v2, the shorter of context timeout or SetTimeout should win
			if err != nil && duration < timeout {
				t.Logf("✓ Driver timeout (%v) was enforced before context timeout (%v)",
					testConfig.SocketTimeout, timeout)
			} else if err == nil {
				t.Logf("✓ Operation completed successfully within both timeouts")
			}
		})
	}
	// Test 2: Long-running operation timeout behavior
	t.Logf("2. Testing long-running operations with driver v2 timeouts")

	// Test with very short timeout
	shortConfig := *config
	shortConfig.SocketTimeout = 100 * time.Millisecond

	shortClient, err := NewClientWithConfig(&shortConfig)
	if err != nil {
		t.Fatalf("Failed to create short timeout client: %v", err)
	}
	defer shortClient.Close()

	shortCollection := shortClient.Collection("timeout_test")
	// Try a find operation that might take longer than 100ms
	ctx := context.Background()
	findStart := time.Now()
	_, err = shortCollection.Find(ctx, map[string]any{})
	findDuration := time.Since(findStart)

	t.Logf("Find with 100ms SocketTimeout took %v, error: %v", findDuration, err)

	// Test 3: Check if timeout configuration is cumulative or individual
	t.Logf("3. Testing timeout configuration precedence")

	// Test the interaction between different timeout types
	complexConfig := *config
	complexConfig.ConnectTimeout = 1 * time.Second      // Very short
	complexConfig.ServerSelectTimeout = 2 * time.Second // Short
	complexConfig.SocketTimeout = 5 * time.Second       // Longer

	t.Logf("Testing with ConnectTimeout=%v, ServerSelectTimeout=%v, SocketTimeout=%v",
		complexConfig.ConnectTimeout, complexConfig.ServerSelectTimeout, complexConfig.SocketTimeout)

	start2 := time.Now()
	complexClient, err := NewClientWithConfig(&complexConfig)
	duration2 := time.Since(start2)

	t.Logf("Client creation with mixed timeouts took %v, error: %v", duration2, err)

	if complexClient != nil {
		defer complexClient.Close()

		// Test an operation
		start3 := time.Now()
		err = complexClient.Ping(context.Background())
		duration3 := time.Since(start3)

		t.Logf("Ping operation took %v, error: %v", duration3, err)
	}
}

func TestTimeoutFailureSimulation(t *testing.T) {
	t.Logf("=== Simulating Timeout Failure Scenarios ===")

	// Test 1: Test connection to non-existent host (should timeout)
	t.Logf("1. Testing connection to non-existent host")

	config := &Config{
		Host:                "non-existent-host-12345.example.com",
		Port:                27017,
		Database:            "test",
		ConnectTimeout:      2 * time.Second,
		ServerSelectTimeout: 2 * time.Second,
		SocketTimeout:       2 * time.Second,
	}

	start := time.Now()
	client, err := NewClientWithConfig(config)
	duration := time.Since(start)

	t.Logf("Connection to non-existent host took %v, error: %v", duration, err)

	if err != nil {
		t.Logf("✓ Expected failure - this is the type of error we see in failing tests")
		t.Logf("   Error type: %T", err)
		if client != nil {
			client.Close()
		}
	}

	// Test 2: Test connection with wrong authentication
	t.Logf("2. Testing connection with wrong authentication")

	wrongAuthConfig := &Config{
		Host:                "localhost",
		Port:                27017,
		Username:            "wrong_user",
		Password:            "wrong_password",
		Database:            "test",
		AuthDatabase:        "admin",
		ConnectTimeout:      5 * time.Second,
		ServerSelectTimeout: 5 * time.Second,
		SocketTimeout:       5 * time.Second,
	}

	start = time.Now()
	wrongClient, err := NewClientWithConfig(wrongAuthConfig)
	duration = time.Since(start)

	t.Logf("Connection with wrong auth took %v, error: %v", duration, err)

	if wrongClient != nil {
		defer wrongClient.Close()

		// Try to ping - this should fail with auth error
		start = time.Now()
		pingErr := wrongClient.Ping(context.Background())
		pingDuration := time.Since(start)

		t.Logf("Ping with wrong auth took %v, error: %v", pingDuration, pingErr)
	}

	// Test 3: Test connection without authentication to auth-required MongoDB
	t.Logf("3. Testing connection without authentication to auth-required MongoDB")

	noAuthConfig := &Config{
		Host:                "localhost",
		Port:                27017,
		Database:            "test",
		ConnectTimeout:      5 * time.Second,
		ServerSelectTimeout: 5 * time.Second,
		SocketTimeout:       5 * time.Second,
	}

	start = time.Now()
	noAuthClient, err := NewClientWithConfig(noAuthConfig)
	duration = time.Since(start)

	t.Logf("Connection without auth took %v, error: %v", duration, err)

	if noAuthClient != nil {
		defer noAuthClient.Close()

		// Try to ping - this might fail or succeed depending on MongoDB config
		start = time.Now()
		pingErr := noAuthClient.Ping(context.Background())
		pingDuration := time.Since(start)

		t.Logf("Ping without auth took %v, error: %v", pingDuration, pingErr)
	}

	// Test 4: Compare with successful connection
	t.Logf("4. Comparing with successful connection")

	goodConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load good config: %v", err)
	}

	start = time.Now()
	goodClient, err := NewClientWithConfig(goodConfig)
	duration = time.Since(start)

	t.Logf("Good connection took %v, error: %v", duration, err)

	if goodClient != nil {
		defer goodClient.Close()

		start = time.Now()
		pingErr := goodClient.Ping(context.Background())
		pingDuration := time.Since(start)

		t.Logf("Good ping took %v, error: %v", pingDuration, pingErr)
	}
}

func TestOriginalTestCaseRecreation(t *testing.T) {
	t.Logf("=== Recreating Original Test Failure Scenario ===")

	// This recreates the exact scenario from TestDatabaseCreation
	// but with detailed logging to understand what went wrong

	t.Logf("1. Testing with default configuration (like original test)")

	// Clear MongoDB environment variables to get defaults
	originalVars := make(map[string]string)
	envVars := []string{
		"MONGODB_HOST", "MONGODB_PORT", "MONGODB_DATABASE",
		"MONGODB_USERNAME", "MONGODB_PASSWORD", "MONGODB_AUTH_DATABASE",
		"MONGODB_SERVER_SELECT_TIMEOUT", "MONGODB_CONNECT_TIMEOUT", "MONGODB_SOCKET_TIMEOUT",
	}

	// Save original values
	for _, env := range envVars {
		originalVars[env] = os.Getenv(env)
	}

	// Test with minimal environment (simulating fresh test environment)
	t.Run("minimal_environment", func(t *testing.T) {
		// Clear all MongoDB environment variables
		for _, env := range envVars {
			os.Unsetenv(env)
		}
		defer func() {
			// Restore original values
			for _, env := range envVars {
				if val, exists := originalVars[env]; exists && val != "" {
					os.Setenv(env, val)
				}
			}
		}()

		t.Logf("Testing NewClient() with no environment variables (defaults only)")

		start := time.Now()
		client, err := NewClient()
		duration := time.Since(start)

		t.Logf("NewClient() with defaults took %v, error: %v", duration, err)

		if err != nil {
			t.Logf("This matches the original test failure!")
			t.Logf("Error type: %T", err)
			t.Logf("Error details: %s", err.Error())
		} else {
			defer client.Close()
			t.Logf("Connection succeeded - checking if it's actually working")

			// Test database creation like original test
			db := client.Database("test_database")
			t.Logf("Database created: %v", db != nil)
		}
	})

	// Test with current environment (should work)
	t.Run("current_environment", func(t *testing.T) {
		t.Logf("Testing NewClient() with current environment variables")

		start := time.Now()
		client, err := NewClient()
		duration := time.Since(start)

		t.Logf("NewClient() with env vars took %v, error: %v", duration, err)

		if client != nil {
			defer client.Close()
		}
	})

	// Test default configuration values
	t.Run("default_config_analysis", func(t *testing.T) {
		// Clear environment temporarily
		for _, env := range envVars {
			os.Unsetenv(env)
		}
		defer func() {
			for _, env := range envVars {
				if val, exists := originalVars[env]; exists && val != "" {
					os.Setenv(env, val)
				}
			}
		}()

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		t.Logf("Default configuration:")
		t.Logf("  Host: %s", config.Host)
		t.Logf("  Port: %d", config.Port)
		t.Logf("  Username: %s", config.Username)
		t.Logf("  Password: %s", config.Password)
		t.Logf("  Database: %s", config.Database)
		t.Logf("  ConnectTimeout: %v", config.ConnectTimeout)
		t.Logf("  ServerSelectTimeout: %v", config.ServerSelectTimeout)
		t.Logf("  SocketTimeout: %v", config.SocketTimeout)

		uri := config.BuildConnectionURI()
		t.Logf("Default URI: %s", uri)
	})
}

func TestEnvironmentBindingDebugging(t *testing.T) {
	t.Logf("=== Testing go-env v1.0.1 with struct tag defaults ===")

	// Clear all MongoDB environment variables first
	envVars := []string{
		"MONGODB_HOST", "MONGODB_PORT", "MONGODB_DATABASE",
		"MONGODB_USERNAME", "MONGODB_PASSWORD", "MONGODB_AUTH_DATABASE",
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}

	// Test 1: Direct env.Bind with empty struct - should apply defaults from struct tags
	t.Logf("1. Testing direct env.Bind with empty struct")

	type TestConfig struct {
		Host string `env:"MONGODB_HOST,default=localhost"`
		Port int    `env:"MONGODB_PORT,default=27017"`
	}

	testConfig := &TestConfig{}
	bindOptions := env.DefaultBindingOptions()

	err := env.Bind(testConfig, bindOptions)
	t.Logf("Direct bind result: Host=%s, Port=%d, Error=%v", testConfig.Host, testConfig.Port, err)

	// Test 2: Test our LoadConfig with empty environment
	t.Logf("2. Testing LoadConfig with no environment variables")

	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Logf("LoadConfig error: %v", err)
	} else {
		t.Logf("LoadConfig result: Host=%s, Port=%d", loadedConfig.Host, loadedConfig.Port)
	}

	// Test 3: Test with some environment variables set
	t.Logf("3. Testing LoadConfig with some environment variables set")

	os.Setenv("MONGODB_HOST", "testhost")
	defer os.Unsetenv("MONGODB_HOST")

	configWithEnv, err := LoadConfig()
	if err != nil {
		t.Logf("LoadConfig with env error: %v", err)
	} else {
		t.Logf("LoadConfig with env result: Host=%s, Port=%d", configWithEnv.Host, configWithEnv.Port)
	}
}

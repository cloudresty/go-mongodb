package mongodb

import (
	"errors"
	"fmt"
	"slices"

	"github.com/cloudresty/go-env"
)

// loadConfigFromEnv loads MongoDB configuration from environment variables (internal function)
func loadConfigFromEnv(prefix string) (*Config, error) {
	// Create empty config struct - go-env v1.0.1 will apply defaults from struct tags
	config := &Config{}

	bindOptions := env.DefaultBindingOptions()
	if prefix != "" {
		bindOptions.Prefix = prefix
	}

	// Bind environment variables and apply defaults from struct tags
	if err := env.Bind(config, bindOptions); err != nil {
		return nil, fmt.Errorf("failed to load environment config: %w", err)
	}

	// Set default logger if none provided
	if config.Logger == nil {
		config.Logger = NopLogger{}
	}

	// Validate the final configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// validateConfig validates the MongoDB configuration
func validateConfig(config *Config) error {
	// Hosts is always required (defaults are set if not provided)
	if config.Hosts == "" {
		return errors.New("MONGODB_HOSTS must be set")
	}

	// All other validation is for enum values only since defaults are set via struct tags
	if !isValidReadPreference(config.ReadPreference) {
		return fmt.Errorf("invalid read preference: %s", config.ReadPreference)
	}

	if !isValidCompressionAlgorithm(config.CompressionAlgorithm) {
		return fmt.Errorf("invalid compression algorithm: %s", config.CompressionAlgorithm)
	}

	if !isValidLogLevel(config.LogLevel) {
		return fmt.Errorf("invalid log level: %s", config.LogLevel)
	}

	if !isValidLogFormat(config.LogFormat) {
		return fmt.Errorf("invalid log format: %s", config.LogFormat)
	}

	if !isValidIDMode(string(config.IDMode)) {
		return fmt.Errorf("invalid ID mode: %s", config.IDMode)
	}

	return nil
}

// isValidReadPreference checks if the read preference is valid
func isValidReadPreference(pref string) bool {
	validPrefs := []string{"primary", "primaryPreferred", "secondary", "secondaryPreferred", "nearest"}
	return slices.Contains(validPrefs, pref)
}

// isValidCompressionAlgorithm checks if the compression algorithm is valid
func isValidCompressionAlgorithm(algo string) bool {
	validAlgos := []string{"snappy", "zlib", "zstd"}
	return slices.Contains(validAlgos, algo)
}

// isValidLogLevel checks if the log level is valid
func isValidLogLevel(level string) bool {
	validLevels := []string{"debug", "info", "warn", "error"}
	return slices.Contains(validLevels, level)
}

// isValidLogFormat checks if the log format is valid
func isValidLogFormat(format string) bool {
	validFormats := []string{"json", "text"}
	return slices.Contains(validFormats, format)
}

// isValidIDMode checks if the ID mode is valid
func isValidIDMode(mode string) bool {
	validModes := []string{"ulid", "objectid", "custom"}
	return slices.Contains(validModes, mode)
}

// Environment variable names for reference
const (
	EnvMongoDBHosts                = "MONGODB_HOSTS"
	EnvMongoDBUsername             = "MONGODB_USERNAME"
	EnvMongoDBPassword             = "MONGODB_PASSWORD"
	EnvMongoDBDatabase             = "MONGODB_DATABASE"
	EnvMongoDBAuthDatabase         = "MONGODB_AUTH_DATABASE"
	EnvMongoDBReplicaSet           = "MONGODB_REPLICA_SET"
	EnvMongoDBMaxPoolSize          = "MONGODB_MAX_POOL_SIZE"
	EnvMongoDBMinPoolSize          = "MONGODB_MIN_POOL_SIZE"
	EnvMongoDBMaxIdleTime          = "MONGODB_MAX_IDLE_TIME"
	EnvMongoDBMaxConnIdleTime      = "MONGODB_MAX_CONN_IDLE_TIME"
	EnvMongoDBConnectTimeout       = "MONGODB_CONNECT_TIMEOUT"
	EnvMongoDBServerSelectTimeout  = "MONGODB_SERVER_SELECT_TIMEOUT"
	EnvMongoDBSocketTimeout        = "MONGODB_SOCKET_TIMEOUT"
	EnvMongoDBHealthCheckEnabled   = "MONGODB_HEALTH_CHECK_ENABLED"
	EnvMongoDBHealthCheckInterval  = "MONGODB_HEALTH_CHECK_INTERVAL"
	EnvMongoDBCompressionEnabled   = "MONGODB_COMPRESSION_ENABLED"
	EnvMongoDBCompressionAlgorithm = "MONGODB_COMPRESSION_ALGORITHM"
	EnvMongoDBReadPreference       = "MONGODB_READ_PREFERENCE"
	EnvMongoDBWriteConcern         = "MONGODB_WRITE_CONCERN"
	EnvMongoDBReadConcern          = "MONGODB_READ_CONCERN"
	EnvMongoDBDirectConnection     = "MONGODB_DIRECT_CONNECTION"
	EnvMongoDBAppName              = "MONGODB_APP_NAME"
	EnvMongoDBConnectionName       = "MONGODB_CONNECTION_NAME"
	EnvMongoDBIDMode               = "MONGODB_ID_MODE"
	EnvMongoDBLogLevel             = "MONGODB_LOG_LEVEL"
	EnvMongoDBLogFormat            = "MONGODB_LOG_FORMAT"
)

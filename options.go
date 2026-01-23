package mongodb

import (
	"crypto/tls"
	"time"

	"go.mongodb.org/mongo-driver/v2/event"
)

// Option represents a functional option for configuring the MongoDB client
type Option func(*Config)

// WithHosts sets the MongoDB host addresses
func WithHosts(hosts ...string) Option {
	return func(c *Config) {
		if len(hosts) > 0 {
			// Join multiple hosts with commas
			hostsStr := ""
			for i, host := range hosts {
				if i > 0 {
					hostsStr += ","
				}
				hostsStr += host
			}
			c.Hosts = hostsStr
		}
	}
}

// WithCredentials sets the authentication credentials
func WithCredentials(username, password string) Option {
	return func(c *Config) {
		c.Username = username
		c.Password = password
	}
}

// WithDatabase sets the default database name
func WithDatabase(name string) Option {
	return func(c *Config) {
		c.Database = name
	}
}

// WithAppName sets the application name for connection metadata
func WithAppName(name string) Option {
	return func(c *Config) {
		c.AppName = name
	}
}

// WithMaxPoolSize sets the maximum number of connections in the pool
func WithMaxPoolSize(size int) Option {
	return func(c *Config) {
		c.MaxPoolSize = uint64(size)
	}
}

// WithMinPoolSize sets the minimum number of connections in the pool
func WithMinPoolSize(size int) Option {
	return func(c *Config) {
		c.MinPoolSize = uint64(size)
	}
}

// WithMaxIdleTime sets the maximum time a connection can remain idle
func WithMaxIdleTime(duration time.Duration) Option {
	return func(c *Config) {
		c.MaxIdleTime = duration
	}
}

// WithTLS enables or disables TLS/SSL.
// When enabled without a custom TLSConfig, uses the system's default TLS configuration.
// For production use with MongoDB Atlas or other secure clusters, this should be enabled.
func WithTLS(enabled bool) Option {
	return func(c *Config) {
		c.TLSEnabled = enabled
	}
}

// WithTLSConfig sets custom TLS configuration.
// This takes precedence over WithTLS(true) and allows fine-grained control over:
//   - Certificate verification (InsecureSkipVerify)
//   - Client certificates (Certificates)
//   - Root CA certificates (RootCAs)
//   - Server name verification (ServerName)
//   - TLS version constraints (MinVersion, MaxVersion)
func WithTLSConfig(config *tls.Config) Option {
	return func(c *Config) {
		c.TLSConfig = config
	}
}

// WithMonitor sets a custom command monitor for APM integration.
// This allows users to plug in monitoring tools like Datadog, OpenTelemetry, etc.
// The monitor receives events for all database commands (find, insert, update, etc.)
// including command start, success, and failure events with timing information.
//
// Example with OpenTelemetry:
//
//	monitor := otelmongo.NewMonitor()
//	client, err := mongodb.NewClient(mongodb.WithMonitor(monitor))
//
// Example with custom logging:
//
//	monitor := &event.CommandMonitor{
//	    Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
//	        log.Printf("Command %s started on %s", evt.CommandName, evt.DatabaseName)
//	    },
//	    Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
//	        log.Printf("Command %s succeeded in %v", evt.CommandName, evt.Duration)
//	    },
//	    Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
//	        log.Printf("Command %s failed: %v", evt.CommandName, evt.Failure)
//	    },
//	}
//	client, err := mongodb.NewClient(mongodb.WithMonitor(monitor))
func WithMonitor(monitor *event.CommandMonitor) Option {
	return func(c *Config) {
		c.CommandMonitor = monitor
	}
}

// WithAuthSource sets the authentication database
func WithAuthSource(source string) Option {
	return func(c *Config) {
		c.AuthDatabase = source
	}
}

// WithReplicaSet sets the replica set name
func WithReplicaSet(name string) Option {
	return func(c *Config) {
		c.ReplicaSet = name
	}
}

// WithReadPreference sets the read preference
func WithReadPreference(pref ReadPreference) Option {
	return func(c *Config) {
		c.ReadPreference = string(pref)
	}
}

// WithWriteConcern sets the write concern
func WithWriteConcern(concern WriteConcern) Option {
	return func(c *Config) {
		c.WriteConcern = string(concern)
	}
}

// WithTimeout sets the default operation timeout
func WithTimeout(duration time.Duration) Option {
	return func(c *Config) {
		c.SocketTimeout = duration
	}
}

// WithConnectTimeout sets the connection timeout
func WithConnectTimeout(duration time.Duration) Option {
	return func(c *Config) {
		c.ConnectTimeout = duration
	}
}

// WithServerSelectionTimeout sets the server selection timeout
func WithServerSelectionTimeout(duration time.Duration) Option {
	return func(c *Config) {
		c.ServerSelectTimeout = duration
	}
}

// WithEnvPrefix sets a custom prefix for environment variables
func WithEnvPrefix(prefix string) Option {
	return func(c *Config) {
		// This would need to be implemented to modify how environment variables are read
		// For now, we'll just note that this is a placeholder
		// The actual implementation would require modifying the env parsing logic
		_ = prefix // Suppress unused variable warning
	}
}

// WithConnectionName sets a local identifier for this client instance
// This is used for application-level logging and metrics, not sent to MongoDB
func WithConnectionName(name string) Option {
	return func(c *Config) {
		c.ConnectionName = name
	}
}

// WithDirectConnection enables or disables direct connection mode
// When enabled, connects directly to a single MongoDB instance without replica set discovery
// Note: This only takes effect when connecting to a single host
func WithDirectConnection(enabled bool) Option {
	return func(c *Config) {
		c.DirectConnection = enabled
	}
}

// WithLogger sets a custom logger implementation for the MongoDB client
// If not provided, the client will use a NopLogger that produces no output
func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// ReadPreference represents MongoDB read preference options
type ReadPreference string

const (
	Primary            ReadPreference = "primary"
	PrimaryPreferred   ReadPreference = "primaryPreferred"
	Secondary          ReadPreference = "secondary"
	SecondaryPreferred ReadPreference = "secondaryPreferred"
	Nearest            ReadPreference = "nearest"
)

// WriteConcern represents MongoDB write concern options
type WriteConcern string

const (
	WCMajority  WriteConcern = "majority"
	WCAcknowl   WriteConcern = "acknowledged"
	WCUnacknowl WriteConcern = "unacknowledged"
	WCJournaled WriteConcern = "journaled"
)

// FromEnv returns an option that loads configuration from environment variables
func FromEnv() Option {
	return func(c *Config) {
		// Load environment variables into the config
		envConfig, err := loadConfigFromEnv("")
		if err == nil {
			*c = *envConfig
		}
	}
}

// FromEnvWithPrefix returns an option that loads configuration from environment variables with a custom prefix
func FromEnvWithPrefix(prefix string) Option {
	return func(c *Config) {
		// Load environment variables with prefix into the config
		envConfig, err := loadConfigFromEnv(prefix)
		if err == nil {
			*c = *envConfig
		}
	}
}

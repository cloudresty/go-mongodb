package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudresty/emit"
	"github.com/cloudresty/ulid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

// IDMode defines the ID generation strategy for documents
type IDMode string

const (
	// IDModeULID generates ULID strings as document IDs (default)
	IDModeULID IDMode = "ulid"
	// IDModeObjectID generates MongoDB ObjectIDs as document IDs
	IDModeObjectID IDMode = "objectid"
	// IDModeCustom allows users to provide their own _id fields
	IDModeCustom IDMode = "custom"
)

// Client represents a MongoDB client with auto-reconnection and environment-first configuration
type Client struct {
	client         *mongo.Client
	config         *Config
	database       *mongo.Database
	mutex          sync.RWMutex
	isConnected    bool
	reconnectCount int64
	lastReconnect  time.Time
	healthTicker   *time.Ticker
	shutdownChan   chan struct{}
	shutdownOnce   sync.Once
}

// Config holds MongoDB connection configuration
type Config struct {
	// Connection settings
	Host         string `env:"MONGODB_HOST,default=localhost"`
	Port         int    `env:"MONGODB_PORT,default=27017"`
	Username     string `env:"MONGODB_USERNAME"`
	Password     string `env:"MONGODB_PASSWORD"`
	Database     string `env:"MONGODB_DATABASE,default=app"`
	AuthDatabase string `env:"MONGODB_AUTH_DATABASE,default=admin"`
	ReplicaSet   string `env:"MONGODB_REPLICA_SET"`

	// Connection pool settings
	MaxPoolSize     uint64        `env:"MONGODB_MAX_POOL_SIZE,default=100"`
	MinPoolSize     uint64        `env:"MONGODB_MIN_POOL_SIZE,default=5"`
	MaxIdleTime     time.Duration `env:"MONGODB_MAX_IDLE_TIME,default=5m"`
	MaxConnIdleTime time.Duration `env:"MONGODB_MAX_CONN_IDLE_TIME,default=10m"`

	// Timeout settings
	ConnectTimeout      time.Duration `env:"MONGODB_CONNECT_TIMEOUT,default=10s"`
	ServerSelectTimeout time.Duration `env:"MONGODB_SERVER_SELECT_TIMEOUT,default=5s"`
	SocketTimeout       time.Duration `env:"MONGODB_SOCKET_TIMEOUT,default=10s"`

	// Reconnection settings
	ReconnectEnabled     bool          `env:"MONGODB_RECONNECT_ENABLED,default=true"`
	ReconnectDelay       time.Duration `env:"MONGODB_RECONNECT_DELAY,default=5s"`
	MaxReconnectDelay    time.Duration `env:"MONGODB_MAX_RECONNECT_DELAY,default=1m"`
	ReconnectBackoff     float64       `env:"MONGODB_RECONNECT_BACKOFF,default=2.0"`
	MaxReconnectAttempts int           `env:"MONGODB_MAX_RECONNECT_ATTEMPTS,default=10"`

	// Health check settings
	HealthCheckEnabled  bool          `env:"MONGODB_HEALTH_CHECK_ENABLED,default=true"`
	HealthCheckInterval time.Duration `env:"MONGODB_HEALTH_CHECK_INTERVAL,default=30s"`

	// Performance settings
	CompressionEnabled   bool   `env:"MONGODB_COMPRESSION_ENABLED,default=true"`
	CompressionAlgorithm string `env:"MONGODB_COMPRESSION_ALGORITHM,default=snappy"`
	ReadPreference       string `env:"MONGODB_READ_PREFERENCE,default=primary"`
	WriteConcern         string `env:"MONGODB_WRITE_CONCERN,default=majority"`
	ReadConcern          string `env:"MONGODB_READ_CONCERN,default=local"`

	// Application settings
	AppName        string `env:"MONGODB_APP_NAME,default=go-mongodb-app"`
	ConnectionName string `env:"MONGODB_CONNECTION_NAME"`

	// ID Generation settings
	IDMode IDMode `env:"MONGODB_ID_MODE,default=ulid"`

	// Logging
	LogLevel  string `env:"MONGODB_LOG_LEVEL,default=info"`
	LogFormat string `env:"MONGODB_LOG_FORMAT,default=json"`
}

// BuildConnectionURI constructs a MongoDB connection URI from configuration components
func (c *Config) BuildConnectionURI() string {
	// Always build URI from components using the priority order:
	// 1. Code defaults 2. Code defaults + env vars 3. All env vars override defaults

	// Build URI from components
	uri := "mongodb://"

	// Add credentials if provided
	if c.Username != "" && c.Password != "" {
		uri += fmt.Sprintf("%s:%s@", c.Username, c.Password)
	}

	// Add host and port
	uri += fmt.Sprintf("%s:%d", c.Host, c.Port)

	// Add database
	if c.Database != "" {
		uri += "/" + c.Database
	}

	// Add query parameters
	params := []string{}

	// Auth database - always include if we have authentication
	if c.Username != "" && c.Password != "" && c.AuthDatabase != "" {
		params = append(params, fmt.Sprintf("authSource=%s", c.AuthDatabase))
	}

	// Replica set
	if c.ReplicaSet != "" {
		params = append(params, fmt.Sprintf("replicaSet=%s", c.ReplicaSet))
	}

	// App name
	if c.AppName != "" {
		params = append(params, fmt.Sprintf("appName=%s", c.AppName))
	}

	// Add compression if enabled
	if c.CompressionEnabled && c.CompressionAlgorithm != "" {
		params = append(params, fmt.Sprintf("compressors=%s", c.CompressionAlgorithm))
	}

	// Add read preference if not default
	if c.ReadPreference != "" && c.ReadPreference != "primary" {
		params = append(params, fmt.Sprintf("readPreference=%s", c.ReadPreference))
	}

	// Add query string if we have parameters
	if len(params) > 0 {
		uri += "?" + joinParams(params)
	}

	return uri
}

// HealthStatus represents the health status of a MongoDB connection
type HealthStatus struct {
	IsHealthy bool          `json:"is_healthy"`
	Error     string        `json:"error,omitempty"`
	Latency   time.Duration `json:"latency"`
	CheckedAt time.Time     `json:"checked_at"`
}

// InsertOneResult represents the result of an insert operation
type InsertOneResult struct {
	InsertedID  string    `json:"inserted_id" bson:"_id"` // ULID used directly as _id
	GeneratedAt time.Time `json:"generated_at" bson:"generated_at"`
}

// InsertManyResult represents the result of a bulk insert operation
type InsertManyResult struct {
	InsertedIDs   []string  `json:"inserted_ids" bson:"inserted_ids"` // ULIDs used directly as _ids
	InsertedCount int64     `json:"inserted_count" bson:"inserted_count"`
	GeneratedAt   time.Time `json:"generated_at" bson:"generated_at"`
}

// UpdateResult represents the result of an update operation
type UpdateResult struct {
	MatchedCount  int64  `json:"matched_count" bson:"matched_count"`
	ModifiedCount int64  `json:"modified_count" bson:"modified_count"`
	UpsertedID    string `json:"upserted_id,omitempty" bson:"upserted_id,omitempty"` // ULID string
	UpsertedCount int64  `json:"upserted_count" bson:"upserted_count"`
}

// DeleteResult represents the result of a delete operation
type DeleteResult struct {
	DeletedCount int64 `json:"deleted_count" bson:"deleted_count"`
}

// QueryOptions provides options for query operations
type QueryOptions struct {
	Sort       bson.D
	Limit      *int64
	Skip       *int64
	Projection bson.D
	Timeout    time.Duration
}

// IndexModel represents a MongoDB index
type IndexModel struct {
	Keys    bson.D
	Options *options.IndexOptionsBuilder
}

// TransactionOptions provides options for transactions
type TransactionOptions struct {
	ReadConcern    *readconcern.ReadConcern
	WriteConcern   *writeconcern.WriteConcern
	ReadPreference *readpref.ReadPref
	MaxCommitTime  *time.Duration
}

// NewClient creates a new MongoDB client with environment-first configuration
func NewClient() (*Client, error) {
	return NewClientWithPrefix("")
}

// NewClientWithPrefix creates a new MongoDB client with a custom environment prefix
func NewClientWithPrefix(prefix string) (*Client, error) {
	config, err := LoadConfigWithPrefix(prefix)
	if err != nil {
		return nil, err
	}

	return NewClientWithConfig(config)
}

// NewClientWithConfig creates a new MongoDB client with the provided configuration
func NewClientWithConfig(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	emit.Info.StructuredFields("Creating new MongoDB client",
		emit.ZString("host", config.Host),
		emit.ZInt("port", config.Port),
		emit.ZString("database", config.Database),
		emit.ZString("app_name", config.AppName))

	client := &Client{
		config:       config,
		shutdownChan: make(chan struct{}),
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if config.HealthCheckEnabled {
		client.startHealthCheck()
	}

	emit.Info.StructuredFields("MongoDB client initialized successfully",
		emit.ZString("host", config.Host),
		emit.ZInt("port", config.Port),
		emit.ZString("database", config.Database),
		emit.ZString("app_name", config.AppName))

	return client, nil
}

// connect establishes a connection to MongoDB
func (c *Client) connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	clientOptions := c.buildClientOptions()

	ctx, cancel := context.WithTimeout(context.Background(), c.config.ConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	c.client = client
	c.database = client.Database(c.config.Database)
	c.isConnected = true
	c.lastReconnect = time.Now()

	return nil
}

// buildClientOptions constructs MongoDB client options from configuration
func (c *Client) buildClientOptions() *options.ClientOptions {
	opts := options.Client()

	// Always build URI from components with priority-based configuration
	uri := c.buildConnectionURI()
	opts.ApplyURI(uri)

	// Connection pool settings
	opts.SetMaxPoolSize(c.config.MaxPoolSize)
	opts.SetMinPoolSize(c.config.MinPoolSize)
	opts.SetMaxConnIdleTime(c.config.MaxConnIdleTime)

	// Timeout settings
	opts.SetConnectTimeout(c.config.ConnectTimeout)
	opts.SetServerSelectionTimeout(c.config.ServerSelectTimeout)
	opts.SetTimeout(c.config.SocketTimeout)

	// Application settings
	opts.SetAppName(c.config.AppName)

	// Compression
	if c.config.CompressionEnabled {
		compressors := []string{c.config.CompressionAlgorithm}
		opts.SetCompressors(compressors)
	}

	// Read preference
	switch c.config.ReadPreference {
	case "primary":
		opts.SetReadPreference(readpref.Primary())
	case "primaryPreferred":
		opts.SetReadPreference(readpref.PrimaryPreferred())
	case "secondary":
		opts.SetReadPreference(readpref.Secondary())
	case "secondaryPreferred":
		opts.SetReadPreference(readpref.SecondaryPreferred())
	case "nearest":
		opts.SetReadPreference(readpref.Nearest())
	}

	return opts
}

// buildConnectionURI constructs a MongoDB connection URI from configuration components
func (c *Client) buildConnectionURI() string {
	uri := "mongodb://"

	if c.config.Username != "" && c.config.Password != "" {
		uri += fmt.Sprintf("%s:%s@", c.config.Username, c.config.Password)
	}

	uri += fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// Add database path if specified
	if c.config.Database != "" {
		uri += "/" + c.config.Database
	}

	params := []string{}

	if c.config.AuthDatabase != "" {
		params = append(params, fmt.Sprintf("authSource=%s", c.config.AuthDatabase))
	}

	if c.config.ReplicaSet != "" {
		params = append(params, fmt.Sprintf("replicaSet=%s", c.config.ReplicaSet))
	}

	if len(params) > 0 {
		uri += "?" + joinParams(params)
	}

	return uri
}

// joinParams joins URL parameters
func joinParams(params []string) string {
	result := ""
	for i, param := range params {
		if i > 0 {
			result += "&"
		}
		result += param
	}
	return result
}

// startHealthCheck starts the health check routine
func (c *Client) startHealthCheck() {
	c.healthTicker = time.NewTicker(c.config.HealthCheckInterval)

	go func() {
		for {
			select {
			case <-c.healthTicker.C:
				c.performHealthCheck()
			case <-c.shutdownChan:
				return
			}
		}
	}()
}

// performHealthCheck checks the health of the MongoDB connection
func (c *Client) performHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	if client == nil {
		emit.Warn.Msg("Health check: client is nil, attempting reconnect")
		go c.attemptReconnect()
		return
	}

	start := time.Now()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		emit.Warn.StructuredFields("Health check failed, attempting reconnect",
			emit.ZString("error", err.Error()))
		c.mutex.Lock()
		c.isConnected = false
		c.mutex.Unlock()

		go c.attemptReconnect()
		return
	}

	latency := time.Since(start)

	emit.Debug.StructuredFields("Health check passed",
		emit.ZDuration("latency", latency))
}

// attemptReconnect attempts to reconnect to MongoDB with exponential backoff
func (c *Client) attemptReconnect() {
	if !c.config.ReconnectEnabled {
		emit.Warn.Msg("Reconnection is disabled")
		return
	}

	c.mutex.RLock()
	isConnected := c.isConnected
	c.mutex.RUnlock()

	if isConnected {
		return // Already connected
	}

	delay := c.config.ReconnectDelay
	maxDelay := c.config.MaxReconnectDelay
	backoff := c.config.ReconnectBackoff
	maxAttempts := c.config.MaxReconnectAttempts

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		emit.Info.StructuredFields("Attempting to reconnect to MongoDB",
			emit.ZInt("attempt", attempt),
			emit.ZInt("max_attempts", maxAttempts),
			emit.ZDuration("delay", delay))

		if err := c.connect(); err != nil {
			emit.Warn.StructuredFields("Reconnection attempt failed",
				emit.ZInt("attempt", attempt),
				emit.ZString("error", err.Error()))

			if attempt < maxAttempts {
				time.Sleep(delay)
				delay = time.Duration(float64(delay) * backoff)
				if delay > maxDelay {
					delay = maxDelay
				}
			}
			continue
		}

		c.mutex.Lock()
		c.reconnectCount++
		c.mutex.Unlock()

		emit.Info.StructuredFields("Successfully reconnected to MongoDB",
			emit.ZInt("attempt", attempt),
			emit.ZInt64("total_reconnects", c.reconnectCount))
		return
	}

	emit.Error.StructuredFields("Failed to reconnect to MongoDB after all attempts",
		emit.ZInt("max_attempts", maxAttempts))
}

// HealthCheck performs a manual health check and returns detailed status
func (c *Client) HealthCheck() *HealthStatus {
	startTime := time.Now()
	status := &HealthStatus{
		CheckedAt: startTime,
	}

	c.mutex.RLock()
	client := c.client
	c.mutex.RUnlock()

	if client == nil {
		status.IsHealthy = false
		status.Error = "client is not connected"
		status.Latency = time.Since(startTime)
		return status
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		status.IsHealthy = false
		status.Error = err.Error()
	} else {
		status.IsHealthy = true
	}

	status.Latency = time.Since(startTime)
	return status
}

// IsConnected returns whether the client is currently connected
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.isConnected
}

// GetReconnectCount returns the number of reconnections performed
func (c *Client) GetReconnectCount() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.reconnectCount
}

// GetLastReconnectTime returns the time of the last reconnection
func (c *Client) GetLastReconnectTime() time.Time {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.lastReconnect
}

// Close gracefully closes the MongoDB connection
func (c *Client) Close() error {
	var closeErr error

	c.shutdownOnce.Do(func() {
		emit.Info.Msg("Closing MongoDB client")

		// Stop health check
		if c.healthTicker != nil {
			c.healthTicker.Stop()
		}

		close(c.shutdownChan)

		c.mutex.Lock()
		defer c.mutex.Unlock()

		if c.client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := c.client.Disconnect(ctx); err != nil {
				emit.Error.StructuredFields("Failed to disconnect MongoDB client",
					emit.ZString("error", err.Error()))
				closeErr = err
			} else {
				emit.Info.Msg("MongoDB client disconnected successfully")
			}
		}

		c.isConnected = false
	})

	return closeErr
}

// generateULID generates a new ULID
func generateULID() string {
	id, _ := ulid.New()
	return id
}

// generateULIDFromTime generates a ULID with a specific timestamp
func generateULIDFromTime(t time.Time) string {
	// Convert time to Unix milliseconds
	timestamp := uint64(t.UnixMilli())
	id, _ := ulid.NewTime(timestamp)
	return id
}

// enhanceDocument adds ID and metadata to a document based on client configuration
func (c *Client) enhanceDocument(doc any) bson.M {
	timestamp := time.Now()

	enhanced := bson.M{
		"created_at": timestamp,
		"updated_at": timestamp,
	}

	// Merge with existing document first
	if docBytes, err := bson.Marshal(doc); err == nil {
		var docMap bson.M
		if err := bson.Unmarshal(docBytes, &docMap); err == nil {
			for k, v := range docMap {
				enhanced[k] = v
			}
		}
	}

	// Generate ID based on client configuration, but only if not already provided
	if _, hasID := enhanced["_id"]; !hasID {
		switch c.config.IDMode {
		case IDModeObjectID:
			enhanced["_id"] = bson.NewObjectID()
		case IDModeCustom:
			// Don't add any _id, let user provide it or let MongoDB auto-generate
		default: // IDModeULID
			enhanced["_id"] = generateULID()
		}
	}

	return enhanced
}

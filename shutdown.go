package mongodb

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ShutdownConfig holds configuration for graceful shutdown
type ShutdownConfig struct {
	Timeout          time.Duration
	GracePeriod      time.Duration
	ForceKillTimeout time.Duration
}

// ShutdownManager manages graceful shutdown of MongoDB connections
type ShutdownManager struct {
	clients          []*Client
	resources        []Shutdownable
	mutex            sync.RWMutex
	shutdownChan     chan os.Signal
	ctx              context.Context
	cancel           context.CancelFunc
	timeout          time.Duration
	gracePeriod      time.Duration
	forceKillTimeout time.Duration
	logger           Logger
}

// Shutdownable interface for resources that can be gracefully shut down
type Shutdownable interface {
	Close() error
}

// NewShutdownManager creates a new shutdown manager
func NewShutdownManager(config *ShutdownConfig) *ShutdownManager {
	if config == nil {
		config = &ShutdownConfig{
			Timeout:          30 * time.Second,
			GracePeriod:      5 * time.Second,
			ForceKillTimeout: 10 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger := NopLogger{} // Default to no-op logger for shutdown manager

	logger.Info("Creating shutdown manager",
		"timeout", config.Timeout,
		"grace_period", config.GracePeriod,
		"force_kill_timeout", config.ForceKillTimeout)

	return &ShutdownManager{
		clients:          make([]*Client, 0),
		resources:        make([]Shutdownable, 0),
		shutdownChan:     make(chan os.Signal, 1),
		ctx:              ctx,
		cancel:           cancel,
		timeout:          config.Timeout,
		gracePeriod:      config.GracePeriod,
		forceKillTimeout: config.ForceKillTimeout,
		logger:           logger,
	}
}

// NewShutdownManagerWithConfig creates a shutdown manager with configuration
func NewShutdownManagerWithConfig(config *Config) *ShutdownManager {
	shutdownConfig := &ShutdownConfig{
		Timeout:          config.ConnectTimeout,
		GracePeriod:      5 * time.Second,
		ForceKillTimeout: 10 * time.Second,
	}

	// Use NopLogger since we don't have access to client logger at this point
	logger := NopLogger{}
	logger.Info("Creating shutdown manager with config",
		"timeout", shutdownConfig.Timeout)

	return NewShutdownManager(shutdownConfig)
}

// Register registers MongoDB clients for graceful shutdown
func (sm *ShutdownManager) Register(clients ...*Client) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.clients = append(sm.clients, clients...)

	sm.logger.Info("Registered clients for graceful shutdown",
		"count", len(clients))
}

// SetupSignalHandler sets up signal handlers for graceful shutdown
func (sm *ShutdownManager) SetupSignalHandler() {
	signal.Notify(sm.shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	sm.logger.Info("Signal handlers setup for graceful shutdown")
}

// Wait blocks until a shutdown signal is received and performs graceful shutdown
func (sm *ShutdownManager) Wait() {
	// Block until signal is received
	sig := <-sm.shutdownChan

	sm.logger.Info("Received shutdown signal, starting graceful shutdown",
		"signal", sig.String())

	sm.shutdown()
}

// Context returns the shutdown manager's context for background workers
func (sm *ShutdownManager) Context() context.Context {
	return sm.ctx
}

// RegisterResources registers shutdownable resources for graceful shutdown
func (sm *ShutdownManager) RegisterResources(resources ...Shutdownable) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.resources = append(sm.resources, resources...)

	sm.logger.Info("Registered shutdownable resources",
		"count", len(resources),
		"total", len(sm.resources))
}

// shutdown performs the actual shutdown logic
func (sm *ShutdownManager) shutdown() {
	// Cancel the context for background workers
	sm.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), sm.timeout)
	defer cancel()

	sm.mutex.RLock()
	clients := make([]*Client, len(sm.clients))
	copy(clients, sm.clients)
	resources := make([]Shutdownable, len(sm.resources))
	copy(resources, sm.resources)
	sm.mutex.RUnlock()

	totalItems := len(clients) + len(resources)
	if totalItems == 0 {
		sm.logger.Info("No clients or resources registered for shutdown")
		return
	}

	// Create a channel to collect shutdown results
	done := make(chan bool, totalItems)
	errorCount := 0

	// Shutdown all clients concurrently
	for i, client := range clients {
		go func(idx int, c *Client) {
			sm.logger.Debug("Shutting down client",
				"index", idx)

			if err := c.Close(); err != nil {
				sm.logger.Error("Failed to close client",
					"index", idx,
					"error", err.Error())
				errorCount++
			} else {
				sm.logger.Debug("Client shut down successfully",
					"index", idx)
			}
			done <- true
		}(i, client)
	}

	// Shutdown all resources concurrently
	for i, resource := range resources {
		go func(idx int, r Shutdownable) {
			sm.logger.Debug("Shutting down resource",
				"index", idx)

			if err := r.Close(); err != nil {
				sm.logger.Error("Failed to close resource",
					"index", idx,
					"error", err.Error())
				errorCount++
			} else {
				sm.logger.Debug("Resource shut down successfully",
					"index", idx)
			}
			done <- true
		}(i, resource)
	}

	// Wait for all items to shutdown or timeout
	shutdownCount := 0
shutdownLoop:
	for shutdownCount < totalItems {
		select {
		case <-done:
			shutdownCount++
		case <-ctx.Done():
			// Timeout reached
			break shutdownLoop
		}
	}

	if shutdownCount == totalItems {
		sm.logger.Info("All clients and resources shut down successfully",
			"count", totalItems)
	} else {
		sm.logger.Warn("Shutdown timeout reached, forcing shutdown",
			"completed", shutdownCount,
			"total", totalItems)
	}

	// Report results
	if errorCount > 0 {
		sm.logger.Warn("Some clients/resources failed to shut down gracefully",
			"errors", errorCount,
			"total", totalItems)
	}

	sm.logger.Info("Graceful shutdown completed")
}

// SetTimeout updates the shutdown timeout
func (sm *ShutdownManager) SetTimeout(timeout time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.timeout = timeout

	sm.logger.Info("Shutdown timeout updated",
		"timeout", timeout)
}

// GetTimeout returns the current shutdown timeout
func (sm *ShutdownManager) GetTimeout() time.Duration {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.timeout
}

// GetClientCount returns the number of registered clients
func (sm *ShutdownManager) GetClientCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return len(sm.clients)
}

// Clear removes all registered clients
func (sm *ShutdownManager) Clear() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.clients = sm.clients[:0]

	sm.logger.Info("Cleared all registered clients",
		"count", 0)
}

// ForceShutdown immediately shuts down all clients without waiting
func (sm *ShutdownManager) ForceShutdown() {
	sm.mutex.RLock()
	clients := make([]*Client, len(sm.clients))
	copy(clients, sm.clients)
	sm.mutex.RUnlock()

	sm.logger.Warn("Performing immediate shutdown",
		"client_count", len(clients))

	for i, client := range clients {
		if err := client.Close(); err != nil {
			sm.logger.Error("Failed to close client during immediate shutdown",
				"index", i,
				"error", err.Error())
		}
	}

	sm.logger.Info("Immediate shutdown completed")
}

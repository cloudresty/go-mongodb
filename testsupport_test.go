package mongodb

import (
	"net"
	"os"
	"testing"
	"time"
)

// skipWithoutMongo skips a test when no MongoDB is listening.
//
// These are INTEGRATION tests: they build a real client and connect. They
// already guard on testing.Short(), but a plain `go test ./...` does not set it,
// so on any machine without MongoDB the package failed — five tests, each after
// a five-second dial timeout.
//
// That matters more than the wasted half-minute. A suite that always fails is a
// suite nobody reads, so a REAL regression here would have been invisible among
// the noise. Skipping keeps the signal.
//
// It probes TCP rather than trusting an environment variable: the point is
// whether a server is actually reachable, and a stale MONGODB_HOSTS pointing at
// nothing would put us right back to a five-second failure.
func skipWithoutMongo(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	address := os.Getenv("MONGODB_TEST_ADDRESS")
	if address == "" {
		address = "localhost:27017"
	}

	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Skipf("skipping: no MongoDB reachable at %s (set MONGODB_TEST_ADDRESS to point elsewhere): %v", address, err)
	}
	_ = conn.Close()
}

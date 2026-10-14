package mongodb

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type logCall struct {
	level string
	msg   string
}

type captureLogger struct{ calls []logCall }

func (l *captureLogger) Info(msg string, _ ...any)  { l.calls = append(l.calls, logCall{"info", msg}) }
func (l *captureLogger) Warn(msg string, _ ...any)  { l.calls = append(l.calls, logCall{"warn", msg}) }
func (l *captureLogger) Error(msg string, _ ...any) { l.calls = append(l.calls, logCall{"error", msg}) }
func (l *captureLogger) Debug(msg string, _ ...any) { l.calls = append(l.calls, logCall{"debug", msg}) }

var (
	dupWriteErr = mongo.WriteException{WriteErrors: mongo.WriteErrors{{Code: 11000, Message: "E11000 duplicate key error"}}}
	dupBulkErr  = mongo.BulkWriteException{WriteErrors: []mongo.BulkWriteError{{WriteError: mongo.WriteError{Code: 11000, Message: "E11000 duplicate key error"}}}}
)

func TestIsExpectedOpError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"duplicate key write exception", dupWriteErr, true},
		{"duplicate key bulk write exception", dupBulkErr, true},
		{"wrapped duplicate key", fmt.Errorf("insert: %w", dupWriteErr), true},
		{"context canceled", context.Canceled, true},
		{"wrapped context canceled", fmt.Errorf("op: %w", context.Canceled), true},
		{"deadline exceeded", context.DeadlineExceeded, false},
		{"other write error", mongo.WriteException{WriteErrors: mongo.WriteErrors{{Code: 121}}}, false},
		{"generic error", errors.New("boom"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExpectedOpError(tt.err); got != tt.want {
				t.Fatalf("isExpectedOpError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestLogOpErrorRouting(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantLevel  string
		wantFailed int64
	}{
		{"duplicate key goes to debug", dupWriteErr, "debug", 0},
		{"canceled goes to debug", context.Canceled, "debug", 0},
		{"deadline stays error", context.DeadlineExceeded, "error", 1},
		{"generic stays error", errors.New("boom"), "error", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg := &captureLogger{}
			c := &Client{config: &Config{Logger: lg}}

			c.logOpError("Failed to insert document", tt.err, "collection", "ledger")
			c.countOpFailure(tt.err)

			if len(lg.calls) != 1 || lg.calls[0].level != tt.wantLevel || lg.calls[0].msg != "Failed to insert document" {
				t.Fatalf("calls = %+v, want one %s call with unchanged message", lg.calls, tt.wantLevel)
			}
			if c.poolStats.operationsFailed != tt.wantFailed {
				t.Fatalf("operationsFailed = %d, want %d", c.poolStats.operationsFailed, tt.wantFailed)
			}
		})
	}
}

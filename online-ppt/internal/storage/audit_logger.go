package storage

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

// AuditLogger writes structured audit events to the provided logger.
type AuditLogger struct {
	logger *log.Logger
	clock  func() time.Time
	mu     sync.Mutex
}

// NewAuditLogger constructs an AuditLogger that emits JSON lines through logger.
// When logger is nil, a default logger writing to stdout is used.
func NewAuditLogger(logger *log.Logger) *AuditLogger {
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}
	return &AuditLogger{
		logger: logger,
		clock:  time.Now,
	}
}

// WithClock overrides the time source. Primarily intended for testing.
func (a *AuditLogger) WithClock(clock func() time.Time) {
	if clock != nil {
		a.clock = clock
	}
}

// Log records an audit event with additional key/value fields encoded as JSON.
func (a *AuditLogger) Log(event string, fields map[string]any) {
	if a == nil {
		return
	}

	payload := make(map[string]any, len(fields)+2)
	if event == "" {
		event = "audit.unknown"
	}
	payload["event"] = event
	payload["timestamp"] = a.clock().UTC().Format(time.RFC3339Nano)
	for k, v := range fields {
		payload[k] = v
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		a.mu.Lock()
		a.logger.Printf("event=audit.encode_failed target=%s error=%v", event, err)
		a.mu.Unlock()
		return
	}

	a.mu.Lock()
	a.logger.Println(string(encoded))
	a.mu.Unlock()
}

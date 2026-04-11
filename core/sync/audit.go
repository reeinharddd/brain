package sync

import (
	"sync"
	"time"
)

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	ID          string
	Timestamp   time.Time
	Operation   string // "push", "pull", "conflict_resolved", "sync"
	Source      string
	Destination string
	ItemsCount  int
	BytesCount  int
	Success     bool
	Error       string
	Details     map[string]string
}

// AuditLogger maintains an append-only audit trail.
type AuditLogger struct {
	mu         sync.RWMutex
	entries    []AuditEntry
	maxEntries int
}

// NewAuditLogger creates a new AuditLogger with the specified maximum number of entries.
func NewAuditLogger(maxEntries int) *AuditLogger {
	return &AuditLogger{
		entries:    make([]AuditEntry, 0, maxEntries),
		maxEntries: maxEntries,
	}
}

// Log appends a new entry to the audit trail.
func (al *AuditLogger) Log(entry AuditEntry) {
	al.mu.Lock()
	defer al.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	al.entries = append(al.entries, entry)

	// Trim oldest entries if we exceed maxEntries
	if len(al.entries) > al.maxEntries {
		al.entries = al.entries[len(al.entries)-al.maxEntries:]
	}
}

// GetEntries returns all audit entries.
func (al *AuditLogger) GetEntries() []AuditEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	result := make([]AuditEntry, len(al.entries))
	copy(result, al.entries)
	return result
}

// GetEntriesSince returns audit entries after the specified time.
func (al *AuditLogger) GetEntriesSince(t time.Time) []AuditEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var result []AuditEntry
	for _, entry := range al.entries {
		if entry.Timestamp.After(t) {
			result = append(result, entry)
		}
	}
	return result
}

// Count returns the number of audit entries.
func (al *AuditLogger) Count() int {
	al.mu.RLock()
	defer al.mu.RUnlock()
	return len(al.entries)
}

// Clear removes all audit entries. Only for testing - in production, never clear.
func (al *AuditLogger) Clear() {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.entries = al.entries[:0]
}

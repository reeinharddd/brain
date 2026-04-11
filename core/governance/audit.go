package governance

import (
	"sync"
	"time"
)

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	ID        string
	Timestamp time.Time
	Action    string            // "policy_resolved", "policy_overridden", "override_requested", "override_approved", "override_denied"
	Subject   string            // who performed the action
	Resource  string            // what was affected
	Details   map[string]string
	Success   bool
}

// AuditLog maintains an append-only log of policy decisions.
type AuditLog struct {
	mu         sync.Mutex
	entries    []AuditEntry
	maxEntries int
}

// NewAuditLog creates a new AuditLog with the given maximum entry count.
func NewAuditLog(maxEntries int) *AuditLog {
	if maxEntries <= 0 {
		maxEntries = 10000
	}
	return &AuditLog{
		entries:    make([]AuditEntry, 0, maxEntries),
		maxEntries: maxEntries,
	}
}

// Log appends an entry to the audit log.
func (al *AuditLog) Log(entry AuditEntry) {
	al.mu.Lock()
	defer al.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	if entry.Details == nil {
		entry.Details = make(map[string]string)
	}

	// If at capacity, remove oldest entry
	if len(al.entries) >= al.maxEntries {
		al.entries = al.entries[1:]
	}

	al.entries = append(al.entries, entry)
}

// GetEntries returns all audit log entries.
func (al *AuditLog) GetEntries() []AuditEntry {
	al.mu.Lock()
	defer al.mu.Unlock()

	result := make([]AuditEntry, len(al.entries))
	copy(result, al.entries)
	return result
}

// GetEntriesByAction returns entries matching the given action.
func (al *AuditLog) GetEntriesByAction(action string) []AuditEntry {
	al.mu.Lock()
	defer al.mu.Unlock()

	var result []AuditEntry
	for _, e := range al.entries {
		if e.Action == action {
			result = append(result, e)
		}
	}
	return result
}

// Count returns the number of entries in the audit log.
func (al *AuditLog) Count() int {
	al.mu.Lock()
	defer al.mu.Unlock()
	return len(al.entries)
}

// Clear removes all entries from the audit log (test only).
func (al *AuditLog) Clear() {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.entries = al.entries[:0]
}

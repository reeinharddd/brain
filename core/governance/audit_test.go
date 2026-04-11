package governance

import (
	"sync"
	"testing"
	"time"
)

func TestAuditLogLog(t *testing.T) {
	al := NewAuditLog(100)

	entry := AuditEntry{
		ID:        "entry-1",
		Timestamp: time.Now().UTC(),
		Action:    "policy_resolved",
		Subject:   "alice",
		Resource:  "rule-1",
		Details:   map[string]string{"scope": "org:acme"},
		Success:   true,
	}

	al.Log(entry)

	if al.Count() != 1 {
		t.Errorf("expected 1 entry, got %d", al.Count())
	}
}

func TestAuditLogGetEntries(t *testing.T) {
	al := NewAuditLog(100)

	al.Log(AuditEntry{ID: "entry-1", Action: "policy_resolved", Subject: "alice"})
	al.Log(AuditEntry{ID: "entry-2", Action: "override_requested", Subject: "bob"})

	entries := al.GetEntries()
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].ID != "entry-1" {
		t.Errorf("expected first entry ID 'entry-1', got %q", entries[0].ID)
	}
	if entries[1].ID != "entry-2" {
		t.Errorf("expected second entry ID 'entry-2', got %q", entries[1].ID)
	}
}

func TestAuditLogGetEntriesByAction(t *testing.T) {
	al := NewAuditLog(100)

	al.Log(AuditEntry{ID: "entry-1", Action: "policy_resolved", Subject: "alice"})
	al.Log(AuditEntry{ID: "entry-2", Action: "override_requested", Subject: "bob"})
	al.Log(AuditEntry{ID: "entry-3", Action: "policy_resolved", Subject: "charlie"})
	al.Log(AuditEntry{ID: "entry-4", Action: "override_approved", Subject: "admin"})

	t.Run("get policy_resolved entries", func(t *testing.T) {
		entries := al.GetEntriesByAction("policy_resolved")
		if len(entries) != 2 {
			t.Errorf("expected 2 policy_resolved entries, got %d", len(entries))
		}
	})

	t.Run("get override_requested entries", func(t *testing.T) {
		entries := al.GetEntriesByAction("override_requested")
		if len(entries) != 1 {
			t.Errorf("expected 1 override_requested entry, got %d", len(entries))
		}
	})

	t.Run("get non-existing action", func(t *testing.T) {
		entries := al.GetEntriesByAction("nonexistent")
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})
}

func TestAuditLogCount(t *testing.T) {
	al := NewAuditLog(100)

	al.Log(AuditEntry{ID: "entry-1"})
	al.Log(AuditEntry{ID: "entry-2"})
	al.Log(AuditEntry{ID: "entry-3"})

	if al.Count() != 3 {
		t.Errorf("expected 3 entries, got %d", al.Count())
	}
}

func TestAuditLogAppendOnly(t *testing.T) {
	al := NewAuditLog(100)

	al.Log(AuditEntry{ID: "entry-1"})
	al.Log(AuditEntry{ID: "entry-2"})

	// GetEntries returns a copy; count remains unchanged
	_ = al.GetEntries()

	if al.Count() != 2 {
		t.Errorf("expected 2 entries, got %d", al.Count())
	}
}

func TestAuditLogMaxEntries(t *testing.T) {
	al := NewAuditLog(5)

	// Add more entries than max
	for i := 0; i < 10; i++ {
		al.Log(AuditEntry{ID: "entry", Action: "test"})
	}

	if al.Count() != 5 {
		t.Errorf("expected 5 entries (max), got %d", al.Count())
	}

	// Oldest entries should be removed
	entries := al.GetEntries()
	// All entries should still be there, just the most recent 5
	if len(entries) != 5 {
		t.Errorf("expected 5 entries in GetEntries, got %d", len(entries))
	}
}

func TestAuditLogClear(t *testing.T) {
	al := NewAuditLog(100)

	al.Log(AuditEntry{ID: "entry-1"})
	al.Log(AuditEntry{ID: "entry-2"})
	al.Log(AuditEntry{ID: "entry-3"})

	al.Clear()

	if al.Count() != 0 {
		t.Errorf("expected 0 entries after clear, got %d", al.Count())
	}
}

func TestAuditLogConcurrentAccess(t *testing.T) {
	al := NewAuditLog(10000)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)

		// Concurrent writes
		go func(i int) {
			defer wg.Done()
			al.Log(AuditEntry{
				ID:     "entry",
				Action: "concurrent_write",
			})
		}(i)

		// Concurrent reads
		go func() {
			defer wg.Done()
			_ = al.GetEntries()
		}()

		go func() {
			defer wg.Done()
			_ = al.GetEntriesByAction("concurrent_write")
		}()
	}

	wg.Wait()

	// Should have 100 entries from the writes
	if al.Count() != 100 {
		t.Errorf("expected 100 entries, got %d", al.Count())
	}
}

func TestAuditLogTimestampAutoSet(t *testing.T) {
	al := NewAuditLog(100)

	entry := AuditEntry{
		ID:     "entry-1",
		Action: "test",
		// No timestamp set
	}

	al.Log(entry)

	entries := al.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].Timestamp.IsZero() {
		t.Error("expected timestamp to be auto-set")
	}
}

func TestAuditLogDetailsNil(t *testing.T) {
	al := NewAuditLog(100)

	entry := AuditEntry{
		ID:      "entry-1",
		Action:  "test",
		Details: nil, // nil details
	}

	al.Log(entry)

	entries := al.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].Details == nil {
		t.Error("expected Details to be initialized to empty map")
	}
}

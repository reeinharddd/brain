package sync

import (
	"testing"
	"time"
)

func TestAuditLogger_Log(t *testing.T) {
	logger := NewAuditLogger(100)

	entry := AuditEntry{
		ID:        "test-1",
		Operation: "push",
		Source:    "local",
		Success:   true,
	}

	logger.Log(entry)

	if logger.Count() != 1 {
		t.Errorf("Count() = %d, want 1", logger.Count())
	}

	entries := logger.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("GetEntries() returned %d entries, want 1", len(entries))
	}

	if entries[0].Operation != "push" {
		t.Errorf("Entry Operation = %s, want push", entries[0].Operation)
	}
}

func TestAuditLogger_GetEntries(t *testing.T) {
	logger := NewAuditLogger(100)

	entries := []AuditEntry{
		{ID: "1", Operation: "push", Success: true},
		{ID: "2", Operation: "pull", Success: true},
		{ID: "3", Operation: "sync", Success: false},
	}

	for _, e := range entries {
		logger.Log(e)
	}

	got := logger.GetEntries()
	if len(got) != len(entries) {
		t.Fatalf("GetEntries() returned %d entries, want %d", len(got), len(entries))
	}

	for i, e := range got {
		if e.ID != entries[i].ID {
			t.Errorf("Entry %d ID = %s, want %s", i, e.ID, entries[i].ID)
		}
		if e.Operation != entries[i].Operation {
			t.Errorf("Entry %d Operation = %s, want %s", i, e.Operation, entries[i].Operation)
		}
	}
}

func TestAuditLogger_GetEntriesSince(t *testing.T) {
	logger := NewAuditLogger(100)

	now := time.Now()

	entries := []AuditEntry{
		{ID: "1", Operation: "push", Timestamp: now.Add(-2 * time.Hour), Success: true},
		{ID: "2", Operation: "pull", Timestamp: now.Add(-1 * time.Hour), Success: true},
		{ID: "3", Operation: "sync", Timestamp: now.Add(-30 * time.Minute), Success: false},
		{ID: "4", Operation: "push", Timestamp: now.Add(-15 * time.Minute), Success: true},
	}

	for _, e := range entries {
		logger.Log(e)
	}

	// Get entries since 45 minutes ago (should get entries 3 and 4)
	since := now.Add(-45 * time.Minute)
	got := logger.GetEntriesSince(since)

	if len(got) != 2 {
		t.Errorf("GetEntriesSince() returned %d entries, want 2", len(got))
	}

	for _, e := range got {
		if !e.Timestamp.After(since) {
			t.Errorf("Entry %s timestamp %v should be after %v", e.ID, e.Timestamp, since)
		}
	}
}

func TestAuditLogger_GetEntriesSince_NoEntries(t *testing.T) {
	logger := NewAuditLogger(100)

	got := logger.GetEntriesSince(time.Now().Add(-time.Hour))
	if len(got) != 0 {
		t.Errorf("GetEntriesSince() on empty logger returned %d entries, want 0", len(got))
	}
}

func TestAuditLogger_Count(t *testing.T) {
	logger := NewAuditLogger(100)

	if logger.Count() != 0 {
		t.Errorf("initial Count() = %d, want 0", logger.Count())
	}

	logger.Log(AuditEntry{ID: "1"})
	if logger.Count() != 1 {
		t.Errorf("Count() after 1 entry = %d, want 1", logger.Count())
	}

	logger.Log(AuditEntry{ID: "2"})
	if logger.Count() != 2 {
		t.Errorf("Count() after 2 entries = %d, want 2", logger.Count())
	}
}

func TestAuditLogger_Clear(t *testing.T) {
	logger := NewAuditLogger(100)

	logger.Log(AuditEntry{ID: "1"})
	logger.Log(AuditEntry{ID: "2"})
	logger.Log(AuditEntry{ID: "3"})

	if logger.Count() != 3 {
		t.Fatalf("Count() before Clear() = %d, want 3", logger.Count())
	}

	logger.Clear()

	if logger.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", logger.Count())
	}
}

func TestAuditLogger_MaxEntries(t *testing.T) {
	maxEntries := 5
	logger := NewAuditLogger(maxEntries)

	// Add more entries than maxEntries
	for i := 0; i < 10; i++ {
		logger.Log(AuditEntry{ID: string(rune('0' + i)), Operation: "push"})
	}

	if logger.Count() > maxEntries {
		t.Errorf("Count() = %d, should be <= %d", logger.Count(), maxEntries)
	}

	// The most recent entries should be kept
	entries := logger.GetEntries()
	// After 10 entries with max 5, we should have entries 5-9
	if len(entries) != maxEntries {
		t.Errorf("GetEntries() returned %d entries, want %d", len(entries), maxEntries)
	}
}

func TestAuditLogger_AppendOnly(t *testing.T) {
	logger := NewAuditLogger(1000)

	// Log some entries
	for i := 0; i < 5; i++ {
		logger.Log(AuditEntry{ID: string(rune('A' + i)), Operation: "push"})
	}

	// Get a snapshot of entries
	initial := logger.GetEntries()
	initialCount := logger.Count()

	// Log more entries
	for i := 0; i < 3; i++ {
		logger.Log(AuditEntry{ID: string(rune('F' + i)), Operation: "pull"})
	}

	// Verify initial entries are still there
	allEntries := logger.GetEntries()
	if len(allEntries) != initialCount+3 {
		t.Errorf("GetEntries() = %d entries, want %d", len(allEntries), initialCount+3)
	}

	// Verify initial entries weren't modified
	for i, e := range initial {
		if allEntries[i].ID != e.ID {
			t.Errorf("Entry %d was modified: got %s, want %s", i, allEntries[i].ID, e.ID)
		}
	}
}

func TestAuditLogger_GetEntriesReturnsCopy(t *testing.T) {
	logger := NewAuditLogger(100)
	logger.Log(AuditEntry{ID: "1", Operation: "push"})

	entries := logger.GetEntries()
	// Modify the returned slice
	if len(entries) > 0 {
		entries[0].Operation = "modified"
	}

	// Original should be unchanged
	original := logger.GetEntries()
	if len(original) > 0 && original[0].Operation == "modified" {
		t.Error("GetEntries() should return a copy, not the original slice")
	}
}

func TestAuditLogger_ConcurrentAccess(t *testing.T) {
	logger := NewAuditLogger(10000)

	done := make(chan bool)

	// Launch multiple goroutines writing
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				logger.Log(AuditEntry{
					ID:        string(rune('A'+id)),
					Operation: "push",
				})
			}
			done <- true
		}(i)
	}

	// Launch readers
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				_ = logger.GetEntries()
				_ = logger.Count()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	if logger.Count() != 1000 {
		t.Errorf("Count() = %d, want 1000", logger.Count())
	}
}

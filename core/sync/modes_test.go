package sync

import (
	"bytes"
	"context"
	"testing"
)

func TestSyncModeHandler_Local(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeLocal, backend, resolver)
	handler := NewSyncModeHandler(ModeLocal, engine)

	ctx := context.Background()
	localData := []byte("local only data")
	remoteData := []byte("remote data")

	result, err := handler.Handle(ctx, localData, remoteData)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !result.Success {
		t.Error("Handle() result.Success should be true")
	}

	if result.Direction != "local" {
		t.Errorf("Direction = %s, want local", result.Direction)
	}

	if result.ItemsSynced != 0 {
		t.Errorf("ItemsSynced = %d, want 0", result.ItemsSynced)
	}

	if result.BytesSynced != len(localData) {
		t.Errorf("BytesSynced = %d, want %d", result.BytesSynced, len(localData))
	}
}

func TestSyncModeHandler_CloudSynced_NoConflict(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)
	handler := NewSyncModeHandler(ModeCloudSynced, engine)

	ctx := context.Background()
	data := []byte("same data")

	result, err := handler.Handle(ctx, data, data)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !result.Success {
		t.Error("Handle() result.Success should be true")
	}

	if result.Direction != "bidirectional" {
		t.Errorf("Direction = %s, want bidirectional", result.Direction)
	}

	if len(result.Conflicts) != 0 {
		t.Errorf("Conflicts = %d, want 0", len(result.Conflicts))
	}
}

func TestSyncModeHandler_CloudSynced_WithConflict(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)
	handler := NewSyncModeHandler(ModeCloudSynced, engine)

	ctx := context.Background()
	localData := []byte("local version")
	remoteData := []byte("remote version")

	result, err := handler.Handle(ctx, localData, remoteData)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !result.Success {
		t.Error("Handle() result.Success should be true")
	}

	if result.Direction != "bidirectional" {
		t.Errorf("Direction = %s, want bidirectional", result.Direction)
	}

	if len(result.Conflicts) != 1 {
		t.Errorf("Conflicts = %d, want 1", len(result.Conflicts))
	}

	conflict := result.Conflicts[0]
	if !bytes.Equal(conflict.LocalData, localData) {
		t.Error("conflict LocalData should match local input")
	}

	if !bytes.Equal(conflict.RemoteData, remoteData) {
		t.Error("conflict RemoteData should match remote input")
	}

	if conflict.ConflictType != "content" {
		t.Errorf("ConflictType = %s, want content", conflict.ConflictType)
	}
}

func TestSyncModeHandler_CloudSynced_EmptyRemote(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)
	handler := NewSyncModeHandler(ModeCloudSynced, engine)

	ctx := context.Background()
	localData := []byte("local data")

	result, err := handler.Handle(ctx, localData, nil)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !result.Success {
		t.Error("Handle() result.Success should be true")
	}

	if len(result.Conflicts) != 0 {
		t.Errorf("Conflicts = %d, want 0 (empty remote is not a conflict)", len(result.Conflicts))
	}
}

func TestSyncModeHandler_Hybrid_NoConflict(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeHybrid, backend, resolver)
	handler := NewSyncModeHandler(ModeHybrid, engine)

	ctx := context.Background()
	data := []byte("same data")

	result, err := handler.Handle(ctx, data, data)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !result.Success {
		t.Error("Handle() result.Success should be true")
	}

	if len(result.Conflicts) != 0 {
		t.Errorf("Conflicts = %d, want 0", len(result.Conflicts))
	}
}

func TestSyncModeHandler_Hybrid_WithConflict(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeHybrid, backend, resolver)
	handler := NewSyncModeHandler(ModeHybrid, engine)

	ctx := context.Background()
	localData := []byte("local version")
	remoteData := []byte("remote version")

	result, err := handler.Handle(ctx, localData, remoteData)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !result.Success {
		t.Error("Handle() result.Success should be true")
	}

	if result.Direction != "bidirectional" {
		t.Errorf("Direction = %s, want bidirectional", result.Direction)
	}

	// Hybrid mode flags conflicts for approval, doesn't auto-resolve
	if len(result.Conflicts) != 1 {
		t.Errorf("Conflicts = %d, want 1", len(result.Conflicts))
	}
}

func TestSyncModeHandler_ContextCancellation(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)
	handler := NewSyncModeHandler(ModeCloudSynced, engine)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := handler.Handle(ctx, []byte("local"), []byte("remote"))
	if err == nil {
		t.Error("Handle() with cancelled context should fail")
	}
}

func TestSyncModeHandler_UnknownMode(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)
	handler := NewSyncModeHandler(SyncMode("unknown"), engine)

	ctx := context.Background()

	_, err := handler.Handle(ctx, []byte("local"), []byte("remote"))
	if err == nil {
		t.Error("Handle() with unknown mode should fail")
	}
}

func TestSyncModeHandler_Hybrid_DoesNotAutoResolve(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeHybrid, backend, resolver)

	ctx := context.Background()
	localData := []byte("local")
	remoteData := []byte("remote")

	// Test through the engine's Sync method
	result, err := engine.Sync(ctx, localData, remoteData)
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if !result.Success {
		t.Error("Sync() result.Success should be true")
	}

	// Hybrid mode should have conflicts that need approval (not resolved)
	if len(result.Conflicts) != 1 {
		t.Errorf("Conflicts = %d, want 1", len(result.Conflicts))
	}
}

func TestSyncModeHandler_CloudSynced_ItemsSynced(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)
	handler := NewSyncModeHandler(ModeCloudSynced, engine)

	ctx := context.Background()

	t.Run("with conflict", func(t *testing.T) {
		result, err := handler.Handle(ctx, []byte("local"), []byte("remote"))
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if result.ItemsSynced != 1 {
			t.Errorf("ItemsSynced = %d, want 1", result.ItemsSynced)
		}
	})

	t.Run("no conflict", func(t *testing.T) {
		data := []byte("same")
		result, err := handler.Handle(ctx, data, data)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if result.ItemsSynced != 0 {
			t.Errorf("ItemsSynced = %d, want 0", result.ItemsSynced)
		}
	})
}

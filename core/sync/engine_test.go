package sync

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// mockBackend is a test implementation of SyncBackend.
type mockBackend struct {
	mu          sync.Mutex
	pushData    []byte
	pushMeta    SyncMetadata
	pullData    []byte
	pullMeta    SyncMetadata
	pushErr     error
	pullErr     error
	healthErr   error
	pushCalled  bool
	pullCalled  bool
}

func (m *mockBackend) Push(ctx context.Context, data []byte, metadata SyncMetadata) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pushCalled = true
	m.pushData = data
	m.pushMeta = metadata
	return m.pushErr
}

func (m *mockBackend) Pull(ctx context.Context) ([]byte, SyncMetadata, error) {
	select {
	case <-ctx.Done():
		return nil, SyncMetadata{}, ctx.Err()
	default:
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pullCalled = true
	return m.pullData, m.pullMeta, m.pullErr
}

func (m *mockBackend) HealthCheck(ctx context.Context) error {
	return m.healthErr
}

func newMockBackend() *mockBackend {
	return &mockBackend{}
}

func TestSyncEngine_Push(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)

	ctx := context.Background()
	data := []byte("test data")
	metadata := SyncMetadata{Version: "v1", Source: "local"}

	result, err := engine.Push(ctx, data, metadata)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if !result.Success {
		t.Error("Push() result.Success should be true")
	}

	if result.Direction != "push" {
		t.Errorf("Direction = %s, want push", result.Direction)
	}

	if result.ItemsSynced != 1 {
		t.Errorf("ItemsSynced = %d, want 1", result.ItemsSynced)
	}

	if !backend.pushCalled {
		t.Error("backend.Push() was not called")
	}
}

func TestSyncEngine_Push_LocalMode(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeLocal, backend, resolver)

	ctx := context.Background()
	data := []byte("test data")
	metadata := SyncMetadata{Version: "v1", Source: "local"}

	result, err := engine.Push(ctx, data, metadata)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if !result.Success {
		t.Error("Push() result.Success should be true")
	}

	if backend.pushCalled {
		t.Error("backend.Push() should not be called in local mode")
	}
}

func TestSyncEngine_Push_Disabled(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)
	engine.Disable()

	ctx := context.Background()
	data := []byte("test data")
	metadata := SyncMetadata{Version: "v1", Source: "local"}

	result, err := engine.Push(ctx, data, metadata)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if result.Success {
		t.Error("Push() result.Success should be false when disabled")
	}

	if result.Error != "sync engine is disabled" {
		t.Errorf("Error = %s, want 'sync engine is disabled'", result.Error)
	}
}

func TestSyncEngine_Push_BackendError(t *testing.T) {
	backend := newMockBackend()
	backend.pushErr = errors.New("push failed")
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)

	ctx := context.Background()
	data := []byte("test data")
	metadata := SyncMetadata{Version: "v1", Source: "local"}

	result, err := engine.Push(ctx, data, metadata)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if result.Success {
		t.Error("Push() result.Success should be false on backend error")
	}

	if result.Error != "push failed" {
		t.Errorf("Error = %s, want 'push failed'", result.Error)
	}
}

func TestSyncEngine_Pull(t *testing.T) {
	backend := newMockBackend()
	backend.pullData = []byte("remote data")
	backend.pullMeta = SyncMetadata{Version: "v2", Source: "remote"}
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)

	ctx := context.Background()

	result, data, meta, err := engine.Pull(ctx)
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	if !result.Success {
		t.Error("Pull() result.Success should be true")
	}

	if result.Direction != "pull" {
		t.Errorf("Direction = %s, want pull", result.Direction)
	}

	if string(data) != "remote data" {
		t.Errorf("data = %s, want 'remote data'", string(data))
	}

	if meta.Version != "v2" {
		t.Errorf("meta.Version = %s, want v2", meta.Version)
	}
}

func TestSyncEngine_Pull_LocalMode(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeLocal, backend, resolver)

	ctx := context.Background()

	result, data, meta, err := engine.Pull(ctx)
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	if !result.Success {
		t.Error("Pull() result.Success should be true")
	}

	if backend.pullCalled {
		t.Error("backend.Pull() should not be called in local mode")
	}

	if data != nil {
		t.Error("data should be nil in local mode")
	}

	if meta.Version != "" {
		t.Error("metadata should be empty in local mode")
	}
}

func TestSyncEngine_Pull_BackendError(t *testing.T) {
	backend := newMockBackend()
	backend.pullErr = errors.New("pull failed")
	backend.pullMeta = SyncMetadata{Source: "remote"}
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)

	ctx := context.Background()

	result, _, _, err := engine.Pull(ctx)
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	if result.Success {
		t.Error("Pull() result.Success should be false on backend error")
	}

	if result.Error != "pull failed" {
		t.Errorf("Error = %s, want 'pull failed'", result.Error)
	}
}

func TestSyncEngine_Sync(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)

	ctx := context.Background()

	result, err := engine.Sync(ctx, []byte("local"), []byte("remote"))
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if !result.Success {
		t.Error("Sync() result.Success should be true")
	}

	if result.Direction != "bidirectional" {
		t.Errorf("Direction = %s, want bidirectional", result.Direction)
	}
}

func TestSyncEngine_Sync_LocalMode(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeLocal, backend, resolver)

	ctx := context.Background()

	result, err := engine.Sync(ctx, []byte("local"), []byte("remote"))
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if !result.Success {
		t.Error("Sync() result.Success should be true")
	}

	if result.Direction != "local" {
		t.Errorf("Direction = %s, want local", result.Direction)
	}
}

func TestSyncEngine_Sync_HybridMode(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeHybrid, backend, resolver)

	ctx := context.Background()

	result, err := engine.Sync(ctx, []byte("local"), []byte("remote"))
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if !result.Success {
		t.Error("Sync() result.Success should be true")
	}

	// Hybrid mode should flag conflicts for approval (not resolve them)
	if len(result.Conflicts) != 1 {
		t.Errorf("Conflicts count = %d, want 1", len(result.Conflicts))
	}
}

func TestSyncEngine_SetMode(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeLocal, backend, resolver)

	if engine.GetMode() != ModeLocal {
		t.Errorf("initial mode = %s, want local", engine.GetMode())
	}

	engine.SetMode(ModeCloudSynced)
	if engine.GetMode() != ModeCloudSynced {
		t.Errorf("mode after SetMode = %s, want cloud-synced", engine.GetMode())
	}

	engine.SetMode(ModeHybrid)
	if engine.GetMode() != ModeHybrid {
		t.Errorf("mode after SetMode(Hybrid) = %s, want hybrid", engine.GetMode())
	}
}

func TestSyncEngine_EnableDisable(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)

	if !engine.IsEnabled() {
		t.Error("engine should be enabled by default")
	}

	engine.Disable()
	if engine.IsEnabled() {
		t.Error("engine should be disabled after Disable()")
	}

	engine.Enable()
	if !engine.IsEnabled() {
		t.Error("engine should be enabled after Enable()")
	}
}

func TestSyncEngine_NoBackend(t *testing.T) {
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, nil, resolver)

	ctx := context.Background()

	_, err := engine.Push(ctx, []byte("data"), SyncMetadata{Source: "local"})
	if err == nil {
		t.Error("Push() with nil backend should fail")
	}

	_, _, _, err = engine.Pull(ctx)
	if err == nil {
		t.Error("Pull() with nil backend should fail")
	}
}

func TestSyncEngine_ContextCancellation(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := engine.Sync(ctx, []byte("local"), []byte("remote"))
	if err == nil {
		t.Error("Sync() with cancelled context should fail")
	}
}

func TestSyncEngine_ConcurrentAccess(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)

	var wg sync.WaitGroup
	ctx := context.Background()

	// Concurrent mode changes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			modes := []SyncMode{ModeLocal, ModeCloudSynced, ModeHybrid}
			engine.SetMode(modes[id%3])
		}(i)
	}

	// Concurrent enable/disable
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				engine.Enable()
			} else {
				engine.Disable()
			}
		}(i)
	}

	// Concurrent push operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			engine.Push(ctx, []byte("data"), SyncMetadata{Source: "local"})
		}()
	}

	wg.Wait()
}

func TestSyncEngine_PushWithEncryption(t *testing.T) {
	backend := newMockBackend()
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)

	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("NewEncryptor() error = %v", err)
	}
	engine.SetEncryptor(encryptor)

	ctx := context.Background()
	data := []byte("secret data")
	metadata := SyncMetadata{Version: "v1", Source: "local"}

	result, err := engine.Push(ctx, data, metadata)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if !result.Success {
		t.Error("Push() result.Success should be true")
	}

	// Backend should have received encrypted data (different from original)
	if string(backend.pushData) == string(data) {
		t.Error("backend should have received encrypted data")
	}
}

func TestSyncEngine_AuditLogging(t *testing.T) {
	backend := newMockBackend()
	backend.pullData = []byte("remote data")
	backend.pullMeta = SyncMetadata{Version: "v1", Source: "remote"}
	resolver := NewConflictResolver(LastWriteWins)
	engine := NewSyncEngine(ModeCloudSynced, backend, resolver)

	ctx := context.Background()

	_, _, _, err := engine.Pull(ctx)
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	logger := engine.GetAuditLogger()
	if logger.Count() == 0 {
		t.Error("audit log should have entries after Pull()")
	}
}

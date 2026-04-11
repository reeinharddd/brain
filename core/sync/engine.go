package sync

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SyncMode defines how sync operates.
type SyncMode string

const (
	ModeLocal       SyncMode = "local"        // No sync, local only
	ModeCloudSynced SyncMode = "cloud-synced" // Automatic cloud sync
	ModeHybrid      SyncMode = "hybrid"       // Explicit approval required
)

// SyncBackend is the interface for sync backends.
type SyncBackend interface {
	Push(ctx context.Context, data []byte, metadata SyncMetadata) error
	Pull(ctx context.Context) ([]byte, SyncMetadata, error)
	HealthCheck(ctx context.Context) error
}

// SyncMetadata holds metadata for sync operations.
type SyncMetadata struct {
	Version   string
	Timestamp time.Time
	Source    string
	Checksum  string
	Encrypted bool
}

// SyncResult holds the result of a sync operation.
type SyncResult struct {
	Success     bool
	Direction   string // "push", "pull", "bidirectional"
	ItemsSynced int
	BytesSynced int
	Conflicts   []Conflict
	Duration    time.Duration
	Error       string
}

// SyncEngine manages memory synchronization.
type SyncEngine struct {
	mu               sync.RWMutex
	mode             SyncMode
	backend          SyncBackend
	conflictResolver *ConflictResolver
	auditLogger      *AuditLogger
	encryptor        *Encryptor
	enabled          bool
}

// NewSyncEngine creates a new SyncEngine.
func NewSyncEngine(mode SyncMode, backend SyncBackend, resolver *ConflictResolver) *SyncEngine {
	return &SyncEngine{
		mode:             mode,
		backend:          backend,
		conflictResolver: resolver,
		auditLogger:      NewAuditLogger(10000),
		enabled:          true,
	}
}

// SetAuditLogger sets a custom audit logger.
func (e *SyncEngine) SetAuditLogger(logger *AuditLogger) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.auditLogger = logger
}

// SetEncryptor sets the encryptor for the sync engine.
func (e *SyncEngine) SetEncryptor(encryptor *Encryptor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.encryptor = encryptor
}

// Push sends data to the remote backend.
func (e *SyncEngine) Push(ctx context.Context, data []byte, metadata SyncMetadata) (*SyncResult, error) {
	start := time.Now()

	e.mu.RLock()
	enabled := e.enabled
	backend := e.backend
	mode := e.mode
	encryptor := e.encryptor
	e.mu.RUnlock()

	if !enabled {
		return &SyncResult{
			Success:  false,
			Error:    "sync engine is disabled",
			Duration: time.Since(start),
		}, nil
	}

	if backend == nil {
		return nil, fmt.Errorf("no backend configured")
	}

	if mode == ModeLocal {
		return &SyncResult{
			Success:     true,
			Direction:   "push",
			ItemsSynced: 0,
			BytesSynced: 0,
			Duration:    time.Since(start),
		}, nil
	}

	// Encrypt data if encryptor is set
	if encryptor != nil {
		encrypted, err := encryptor.Encrypt(data)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt data: %w", err)
		}
		data = encrypted
		metadata.Encrypted = true
	}

	metadata.Timestamp = time.Now()

	err := backend.Push(ctx, data, metadata)
	if err != nil {
		e.logAudit(AuditEntry{
			Operation:   "push",
			Source:      "local",
			Destination: metadata.Source,
			BytesCount:  len(data),
			Success:     false,
			Error:       err.Error(),
		})
		return &SyncResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start),
		}, nil
	}

	e.logAudit(AuditEntry{
		Operation:   "push",
		Source:      "local",
		Destination: metadata.Source,
		BytesCount:  len(data),
		Success:     true,
	})

	return &SyncResult{
		Success:     true,
		Direction:   "push",
		ItemsSynced: 1,
		BytesSynced: len(data),
		Duration:    time.Since(start),
	}, nil
}

// Pull retrieves data from the remote backend.
func (e *SyncEngine) Pull(ctx context.Context) (*SyncResult, []byte, SyncMetadata, error) {
	start := time.Now()

	e.mu.RLock()
	enabled := e.enabled
	backend := e.backend
	mode := e.mode
	encryptor := e.encryptor
	e.mu.RUnlock()

	if !enabled {
		return &SyncResult{
			Success:  false,
			Error:    "sync engine is disabled",
			Duration: time.Since(start),
		}, nil, SyncMetadata{}, nil
	}

	if backend == nil {
		return nil, nil, SyncMetadata{}, fmt.Errorf("no backend configured")
	}

	if mode == ModeLocal {
		return &SyncResult{
			Success:     true,
			Direction:   "pull",
			ItemsSynced: 0,
			BytesSynced: 0,
			Duration:    time.Since(start),
		}, nil, SyncMetadata{}, nil
	}

	data, metadata, err := backend.Pull(ctx)
	if err != nil {
		e.logAudit(AuditEntry{
			Operation: "pull",
			Source:    metadata.Source,
			Destination: "local",
			Success:   false,
			Error:     err.Error(),
		})
		return &SyncResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start),
		}, nil, metadata, nil
	}

	// Decrypt data if encryptor is set and data is encrypted
	if encryptor != nil && metadata.Encrypted {
		decrypted, err := encryptor.Decrypt(data)
		if err != nil {
			return nil, nil, metadata, fmt.Errorf("failed to decrypt data: %w", err)
		}
		data = decrypted
	}

	e.logAudit(AuditEntry{
		Operation:   "pull",
		Source:      metadata.Source,
		Destination: "local",
		BytesCount:  len(data),
		Success:     true,
	})

	return &SyncResult{
		Success:     true,
		Direction:   "pull",
		ItemsSynced: 1,
		BytesSynced: len(data),
		Duration:    time.Since(start),
	}, data, metadata, nil
}

// Sync performs bidirectional synchronization between local and remote data.
func (e *SyncEngine) Sync(ctx context.Context, localData, remoteData []byte) (*SyncResult, error) {
	start := time.Now()

	e.mu.RLock()
	enabled := e.enabled
	mode := e.mode
	resolver := e.conflictResolver
	e.mu.RUnlock()

	if !enabled {
		return &SyncResult{
			Success:  false,
			Error:    "sync engine is disabled",
			Duration: time.Since(start),
		}, nil
	}

	handler := NewSyncModeHandler(mode, e)
	result, err := handler.Handle(ctx, localData, remoteData)
	if err != nil {
		return nil, fmt.Errorf("sync failed: %w", err)
	}
	result.Duration = time.Since(start)

	// Resolve conflicts if any
	if len(result.Conflicts) > 0 && resolver != nil && mode != ModeHybrid {
		var resolved []Conflict
		for _, conflict := range result.Conflicts {
			resolution, err := resolver.Resolve(ctx, conflict, "default")
			if err != nil {
				result.Error = fmt.Sprintf("conflict resolution failed: %v", err)
				result.Success = false
				return result, nil
			}
			if resolution.RequiresApproval {
				resolved = append(resolved, conflict)
			}
		}
		result.Conflicts = resolved
	}

	e.logAudit(AuditEntry{
		Operation:   "sync",
		Source:      "local",
		Destination: "remote",
		ItemsCount:  result.ItemsSynced,
		BytesCount:  result.BytesSynced,
		Success:     result.Success,
		Error:       result.Error,
	})

	return result, nil
}

// SetMode changes the sync mode.
func (e *SyncEngine) SetMode(mode SyncMode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.mode = mode
}

// GetMode returns the current sync mode.
func (e *SyncEngine) GetMode() SyncMode {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.mode
}

// Enable enables the sync engine.
func (e *SyncEngine) Enable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = true
}

// Disable disables the sync engine.
func (e *SyncEngine) Disable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = false
}

// IsEnabled returns whether the sync engine is enabled.
func (e *SyncEngine) IsEnabled() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enabled
}

// logAudit safely logs an audit entry.
func (e *SyncEngine) logAudit(entry AuditEntry) {
	e.mu.RLock()
	logger := e.auditLogger
	e.mu.RUnlock()

	if logger != nil {
		logger.Log(entry)
	}
}

// GetAuditLogger returns the audit logger.
func (e *SyncEngine) GetAuditLogger() *AuditLogger {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.auditLogger
}

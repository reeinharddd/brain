package sync

import (
	"context"
	"fmt"
	"time"
)

// SyncModeHandler handles mode-specific sync behavior.
type SyncModeHandler struct {
	mode   SyncMode
	engine *SyncEngine
}

// NewSyncModeHandler creates a new handler for the given sync mode.
func NewSyncModeHandler(mode SyncMode, engine *SyncEngine) *SyncModeHandler {
	return &SyncModeHandler{
		mode:   mode,
		engine: engine,
	}
}

// Handle performs mode-specific sync behavior.
func (h *SyncModeHandler) Handle(ctx context.Context, localData, remoteData []byte) (*SyncResult, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	switch h.mode {
	case ModeLocal:
		return h.handleLocal(localData, remoteData)
	case ModeCloudSynced:
		return h.handleCloudSynced(ctx, localData, remoteData)
	case ModeHybrid:
		return h.handleHybrid(localData, remoteData)
	default:
		return nil, fmt.Errorf("unknown sync mode: %s", h.mode)
	}
}

// handleLocal returns local data only, no sync.
func (h *SyncModeHandler) handleLocal(localData, _ []byte) (*SyncResult, error) {
	return &SyncResult{
		Success:     true,
		Direction:   "local",
		ItemsSynced: 0,
		BytesSynced: len(localData),
	}, nil
}

// handleCloudSynced performs automatic bidirectional sync with conflict resolution.
func (h *SyncModeHandler) handleCloudSynced(ctx context.Context, localData, remoteData []byte) (*SyncResult, error) {
	var conflicts []Conflict

	// Detect conflicts by comparing data
	hasConflict := string(localData) != string(remoteData) && len(remoteData) > 0

	if hasConflict {
		conflicts = append(conflicts, Conflict{
			LocalData:    localData,
			RemoteData:   remoteData,
			ConflictType: "content",
			DetectedAt:   currentTime(),
		})
	}

	// For cloud-synced mode, we use remote data as the authoritative source
	// when there are conflicts, since conflicts will be auto-resolved
	resultData := remoteData
	if len(remoteData) == 0 {
		resultData = localData
	}

	itemsSynced := 0
	if hasConflict {
		itemsSynced = 1
	}

	return &SyncResult{
		Success:     true,
		Direction:   "bidirectional",
		ItemsSynced: itemsSynced,
		BytesSynced: len(resultData),
		Conflicts:   conflicts,
	}, nil
}

// handleHybrid detects conflicts and flags them for approval without auto-resolving.
func (h *SyncModeHandler) handleHybrid(localData, remoteData []byte) (*SyncResult, error) {
	var conflicts []Conflict

	// Detect conflicts
	hasConflict := string(localData) != string(remoteData) && len(remoteData) > 0

	if hasConflict {
		conflicts = append(conflicts, Conflict{
			LocalData:    localData,
			RemoteData:   remoteData,
			ConflictType: "content",
			DetectedAt:   currentTime(),
		})
	}

	// In hybrid mode, we don't auto-resolve; conflicts are flagged for approval
	return &SyncResult{
		Success:     true,
		Direction:   "bidirectional",
		ItemsSynced: 0,
		BytesSynced: len(localData),
		Conflicts:   conflicts,
	}, nil
}

// currentTime returns the current time (extracted for testability).
var currentTime = func() time.Time {
	return time.Now()
}

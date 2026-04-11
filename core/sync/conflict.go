package sync

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ConflictResolutionStrategy defines how conflicts are resolved.
type ConflictResolutionStrategy string

const (
	LastWriteWins   ConflictResolutionStrategy = "last_write_wins"
	Manual          ConflictResolutionStrategy = "manual"
	Merge           ConflictResolutionStrategy = "merge"
	LocalPreferred  ConflictResolutionStrategy = "local_preferred"
	RemotePreferred ConflictResolutionStrategy = "remote_preferred"
)

// Conflict represents a detected conflict.
type Conflict struct {
	LocalData     []byte
	RemoteData    []byte
	ConflictType  string // "content", "metadata", "state"
	LocalVersion  string
	RemoteVersion string
	DetectedAt    time.Time
}

// Resolution represents a conflict resolution.
type Resolution struct {
	Conflict         Conflict
	Strategy         ConflictResolutionStrategy
	Winner           string // "local", "remote", "merged"
	MergedData       []byte
	ResolvedAt       time.Time
	RequiresApproval bool
}

// ConflictResolver resolves conflicts between local and remote data.
type ConflictResolver struct {
	mu               sync.RWMutex
	defaultStrategy  ConflictResolutionStrategy
	typeStrategies   map[string]ConflictResolutionStrategy
	auditLogger      *AuditLogger
}

// NewConflictResolver creates a new ConflictResolver with the given default strategy.
func NewConflictResolver(defaultStrategy ConflictResolutionStrategy) *ConflictResolver {
	return &ConflictResolver{
		defaultStrategy: defaultStrategy,
		typeStrategies:  make(map[string]ConflictResolutionStrategy),
	}
}

// SetTypeStrategy sets a conflict resolution strategy for a specific artifact type.
func (cr *ConflictResolver) SetTypeStrategy(artifactType string, strategy ConflictResolutionStrategy) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.typeStrategies[artifactType] = strategy
}

// Resolve resolves a conflict using the appropriate strategy.
func (cr *ConflictResolver) Resolve(ctx context.Context, conflict Conflict, artifactType string) (*Resolution, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled during conflict resolution: %w", ctx.Err())
	default:
	}

	strategy := cr.getStrategy(artifactType)

	var resolution *Resolution
	switch strategy {
	case LastWriteWins:
		resolution = cr.lastWriteWins(conflict)
	case LocalPreferred:
		resolution = cr.localPreferred(conflict)
	case RemotePreferred:
		resolution = cr.remotePreferred(conflict)
	case Merge:
		resolution = cr.merge(conflict)
	case Manual:
		resolution = cr.manual(conflict)
	default:
		return nil, fmt.Errorf("unknown conflict resolution strategy: %s", strategy)
	}

	resolution.Strategy = strategy
	resolution.ResolvedAt = time.Now()

	// Log to audit trail if audit logger is set
	if cr.auditLogger != nil {
		cr.auditLogger.Log(AuditEntry{
			Operation: "conflict_resolved",
			Source:    "conflict_resolver",
			Details: map[string]string{
				"strategy": string(strategy),
				"winner":   resolution.Winner,
				"type":     artifactType,
			},
		})
	}

	return resolution, nil
}

// getStrategy returns the strategy for the given artifact type, falling back to default.
func (cr *ConflictResolver) getStrategy(artifactType string) ConflictResolutionStrategy {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	if strategy, ok := cr.typeStrategies[artifactType]; ok {
		return strategy
	}
	return cr.defaultStrategy
}

// lastWriteWins resolves conflicts by selecting the most recently modified version.
func (cr *ConflictResolver) lastWriteWins(conflict Conflict) *Resolution {
	// Compare versions as a proxy for recency; if versions can't be compared, prefer remote
	winner := "remote"
	winnerData := conflict.RemoteData

	// Simple comparison: if local version is "greater" (lexicographically), use local
	if conflict.LocalVersion > conflict.RemoteVersion {
		winner = "local"
		winnerData = conflict.LocalData
	}

	return &Resolution{
		Conflict:         conflict,
		Winner:           winner,
		MergedData:       winnerData,
		RequiresApproval: false,
	}
}

// localPreferred resolves conflicts by selecting the local version.
func (cr *ConflictResolver) localPreferred(conflict Conflict) *Resolution {
	return &Resolution{
		Conflict:         conflict,
		Winner:           "local",
		MergedData:       conflict.LocalData,
		RequiresApproval: false,
	}
}

// remotePreferred resolves conflicts by selecting the remote version.
func (cr *ConflictResolver) remotePreferred(conflict Conflict) *Resolution {
	return &Resolution{
		Conflict:         conflict,
		Winner:           "remote",
		MergedData:       conflict.RemoteData,
		RequiresApproval: false,
	}
}

// merge resolves conflicts by concatenating local and remote data with a separator.
func (cr *ConflictResolver) merge(conflict Conflict) *Resolution {
	merged := append(conflict.LocalData, []byte("\n--- CONFLICT MERGE BOUNDARY ---\n")...)
	merged = append(merged, conflict.RemoteData...)

	return &Resolution{
		Conflict:         conflict,
		Winner:           "merged",
		MergedData:       merged,
		RequiresApproval: false,
	}
}

// manual resolves conflicts by flagging them for approval.
func (cr *ConflictResolver) manual(conflict Conflict) *Resolution {
	return &Resolution{
		Conflict:         conflict,
		Winner:           "pending",
		MergedData:       nil,
		RequiresApproval: true,
	}
}

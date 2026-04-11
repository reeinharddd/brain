package autoevolve

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// UsageTelemetry captures a single usage event
type UsageTelemetry struct {
	Timestamp        time.Time
	Surface          string // which IDE/CLI
	ActionType       string // skill_used, mcp_called, policy_checked, ...
	ArtifactKind     string // skill, mcp, rule, agent
	ArtifactID       string
	Success          bool
	Duration         time.Duration
	TokensUsed       int
	TokensWasted     int // Unused context, failed retries
	ErrorType        string
	UserSatisfaction *int // 1-5 if surface provides feedback
}

// TelemetryAccumulator collects telemetry events
type TelemetryAccumulator struct {
	mu        sync.RWMutex
	events    []UsageTelemetry
	maxEvents int
	// Tracking maps
	failureTracker map[string]*FailureStats // artifactID -> stats
	gapDetector    map[string]*GapStats     // searchQuery -> stats
	wasteTracker   map[string]*WasteStats   // surface -> stats
}

// FailureStats tracks failure patterns
type FailureStats struct {
	TotalAttempts int
	Failures      int
	LastFailure   time.Time
	ErrorTypes    map[string]int
	FailureRate   float64
}

// GapStats tracks missing artifact searches
type GapStats struct {
	Query       string
	SearchCount int
	TopSurfaces map[string]int
	FirstSeen   time.Time
	LastSeen    time.Time
}

// WasteStats tracks token waste
type WasteStats struct {
	TotalWasted int
	TotalUsed   int
	WasteRate   float64 // wasted/used ratio
	Sessions    int
}

func NewTelemetryAccumulator(maxEvents int) *TelemetryAccumulator {
	return &TelemetryAccumulator{
		events:         make([]UsageTelemetry, 0, maxEvents),
		maxEvents:      maxEvents,
		failureTracker: make(map[string]*FailureStats),
		gapDetector:    make(map[string]*GapStats),
		wasteTracker:   make(map[string]*WasteStats),
	}
}

func (a *TelemetryAccumulator) Record(ctx context.Context, event UsageTelemetry) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("telemetry record: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Enforce maxEvents limit by dropping oldest
	if len(a.events) >= a.maxEvents {
		a.events = a.events[1:]
	}
	a.events = append(a.events, event)

	// Update failure tracking
	if event.ArtifactID != "" {
		stats, ok := a.failureTracker[event.ArtifactID]
		if !ok {
			stats = &FailureStats{
				ErrorTypes: make(map[string]int),
			}
			a.failureTracker[event.ArtifactID] = stats
		}
		stats.TotalAttempts++
		if !event.Success {
			stats.Failures++
			stats.LastFailure = event.Timestamp
			if event.ErrorType != "" {
				stats.ErrorTypes[event.ErrorType]++
			}
		}
		if stats.TotalAttempts > 0 {
			stats.FailureRate = float64(stats.Failures) / float64(stats.TotalAttempts)
		}
	}

	// Update gap detection (action types that indicate a search for something not found)
	if event.ActionType == "search" && !event.Success && event.ArtifactID != "" {
		stats, ok := a.gapDetector[event.ArtifactID]
		if !ok {
			stats = &GapStats{
				Query:       event.ArtifactID,
				TopSurfaces: make(map[string]int),
				FirstSeen:   event.Timestamp,
			}
			a.gapDetector[event.ArtifactID] = stats
		}
		stats.SearchCount++
		stats.LastSeen = event.Timestamp
		if event.Surface != "" {
			stats.TopSurfaces[event.Surface]++
		}
	}

	// Update waste tracking
	if event.Surface != "" {
		stats, ok := a.wasteTracker[event.Surface]
		if !ok {
			stats = &WasteStats{}
			a.wasteTracker[event.Surface] = stats
		}
		stats.TotalWasted += event.TokensWasted
		stats.TotalUsed += event.TokensUsed
		stats.Sessions++
		if stats.TotalUsed > 0 {
			stats.WasteRate = float64(stats.TotalWasted) / float64(stats.TotalUsed)
		}
	}

	return nil
}

func (a *TelemetryAccumulator) GetEvents(ctx context.Context) []UsageTelemetry {
	if err := ctx.Err(); nil != err {
		return nil
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]UsageTelemetry, len(a.events))
	copy(result, a.events)
	return result
}

func (a *TelemetryAccumulator) GetFailureStats(ctx context.Context) map[string]*FailureStats {
	if err := ctx.Err(); err != nil {
		return nil
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]*FailureStats, len(a.failureTracker))
	for k, v := range a.failureTracker {
		// Deep copy
		etCopy := make(map[string]int, len(v.ErrorTypes))
		for ek, ev := range v.ErrorTypes {
			etCopy[ek] = ev
		}
		result[k] = &FailureStats{
			TotalAttempts: v.TotalAttempts,
			Failures:      v.Failures,
			LastFailure:   v.LastFailure,
			ErrorTypes:    etCopy,
			FailureRate:   v.FailureRate,
		}
	}
	return result
}

func (a *TelemetryAccumulator) GetGapStats(ctx context.Context) map[string]*GapStats {
	if err := ctx.Err(); err != nil {
		return nil
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]*GapStats, len(a.gapDetector))
	for k, v := range a.gapDetector {
		surfaceCopy := make(map[string]int, len(v.TopSurfaces))
		for sk, sv := range v.TopSurfaces {
			surfaceCopy[sk] = sv
		}
		result[k] = &GapStats{
			Query:       v.Query,
			SearchCount: v.SearchCount,
			TopSurfaces: surfaceCopy,
			FirstSeen:   v.FirstSeen,
			LastSeen:    v.LastSeen,
		}
	}
	return result
}

func (a *TelemetryAccumulator) GetWasteStats(ctx context.Context) map[string]*WasteStats {
	if err := ctx.Err(); err != nil {
		return nil
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]*WasteStats, len(a.wasteTracker))
	for k, v := range a.wasteTracker {
		result[k] = &WasteStats{
			TotalWasted: v.TotalWasted,
			TotalUsed:   v.TotalUsed,
			WasteRate:   v.WasteRate,
			Sessions:    v.Sessions,
		}
	}
	return result
}

func (a *TelemetryAccumulator) Count() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.events)
}

// Clear resets all accumulated data. Intended for test use only.
func (a *TelemetryAccumulator) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = a.events[:0]
	a.failureTracker = make(map[string]*FailureStats)
	a.gapDetector = make(map[string]*GapStats)
	a.wasteTracker = make(map[string]*WasteStats)
}

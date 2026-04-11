package curator

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AutoDreamService manages idle-time context consolidation
type AutoDreamService struct {
	config          CuratorConfig
	mu              sync.RWMutex
	lastDreamTime   time.Time
	sessionCounter  int
	curator         *CuratorService
	idleDeadline    time.Duration
	minHoursDream   time.Duration
	minSessions     int
}

// NewAutoDreamService creates a new auto-dream service
func NewAutoDreamService(cfg CuratorConfig, idleDeadline time.Duration) *AutoDreamService {
	return &AutoDreamService{
		config:        cfg,
		curator:       NewCuratorService(cfg),
		idleDeadline:  idleDeadline,
		minHoursDream: 24 * time.Hour, // Default: dream at most once per day
		minSessions:   3,              // Default: need at least 3 sessions
	}
}

// RecordSession increments the session counter
func (ads *AutoDreamService) RecordSession(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("record session cancelled: %w", ctx.Err())
	default:
	}

	ads.mu.Lock()
	defer ads.mu.Unlock()

	ads.sessionCounter++
	return nil
}

// Tick evaluates whether to trigger a dream cycle
func (ads *AutoDreamService) Tick(ctx context.Context, contentEntries map[string]string, accessRecords map[string]AccessRecord) (*CuratorReport, error) {
	ads.mu.Lock()

	timeSinceDream := time.Since(ads.lastDreamTime)
	readyToDream := timeSinceDream >= ads.minHoursDream
	enoughSessions := ads.sessionCounter >= ads.minSessions

	if !readyToDream || !enoughSessions {
		ads.mu.Unlock()
		return nil, nil
	}

	// Reset counters
	ads.sessionCounter = 0
	ads.lastDreamTime = time.Now()
	ads.mu.Unlock()

	// Run curator
	report, err := ads.curator.Run(ctx, contentEntries, accessRecords)
	if err != nil {
		return nil, fmt.Errorf("dream curation failed: %w", err)
	}

	return report, nil
}

// ShouldTrigger checks if conditions are met for a dream cycle
func (ads *AutoDreamService) ShouldTrigger() bool {
	ads.mu.RLock()
	defer ads.mu.RUnlock()

	timeSinceDream := time.Since(ads.lastDreamTime)
	return timeSinceDream >= ads.minHoursDream && ads.sessionCounter >= ads.minSessions
}

// GetStatus returns the current status of the auto-dream service
func (ads *AutoDreamService) GetStatus() map[string]interface{} {
	ads.mu.RLock()
	defer ads.mu.RUnlock()

	return map[string]interface{}{
		"last_dream_time":  ads.lastDreamTime,
		"session_counter":  ads.sessionCounter,
		"time_since_dream": time.Since(ads.lastDreamTime),
		"ready_to_dream":   time.Since(ads.lastDreamTime) >= ads.minHoursDream,
		"enough_sessions":  ads.sessionCounter >= ads.minSessions,
		"should_trigger":   time.Since(ads.lastDreamTime) >= ads.minHoursDream && ads.sessionCounter >= ads.minSessions,
	}
}

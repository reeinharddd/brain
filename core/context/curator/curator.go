package curator

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// CuratorConfig holds curator settings
type CuratorConfig struct {
	Enabled            bool
	DryRun             bool
	DedupThreshold     float64
	MaxTokensPerLayer  int
	MinPromotionAccess int
	PromotionRecency   time.Duration
	StaleThreshold     time.Duration
	MinCleanupAccess   int
}

// MemoryState describes current memory health
type MemoryState struct {
	TotalEntries   int
	TotalTokens    int
	DuplicateCount int
	OversizedCount int
	StaleCount     int
	HealthScore    float64 // 0.0-1.0, higher is better
}

// CuratorReport is the full analysis result
type CuratorReport struct {
	RunAt       time.Time
	DryRun      bool
	Duplicates  []DuplicateFinding
	Compactions []CompactionSuggestion
	Promotions  []PromotionSuggestion
	Cleanups    []CleanupSuggestion
	MemoryState MemoryState
	TokenSavings int
}

// CuratorService orchestrates context curation
type CuratorService struct {
	config    CuratorConfig
	detector  *DeduplicationDetector
	compactor *Compactor
	promoter  *Promoter
	cleaner   *CleanupAdvisor
}

// NewCuratorService creates a new curator service
func NewCuratorService(cfg CuratorConfig) *CuratorService {
	return &CuratorService{
		config:    cfg,
		detector:  NewDeduplicationDetector(cfg.DedupThreshold),
		compactor: NewCompactor(cfg.MaxTokensPerLayer),
		promoter:  NewPromoter(cfg.MinPromotionAccess, cfg.PromotionRecency),
		cleaner:   NewCleanupAdvisor(cfg.StaleThreshold, cfg.MinCleanupAccess),
	}
}

// Run executes the full curation pipeline
func (c *CuratorService) Run(ctx context.Context, contentEntries map[string]string, accessRecords map[string]AccessRecord) (*CuratorReport, error) {
	if !c.config.Enabled {
		return nil, errors.New("curator service is not enabled")
	}

	// Step 2: Run deduplication detection
	duplicates, err := c.detector.Detect(ctx, contentEntries)
	if err != nil {
		return nil, fmt.Errorf("deduplication failed: %w", err)
	}

	// Step 3: Run compaction analysis
	compactions, err := c.compactor.Analyze(ctx, contentEntries)
	if err != nil {
		return nil, fmt.Errorf("compaction analysis failed: %w", err)
	}

	// Step 4: Run promotion analysis
	promotions, err := c.promoter.Analyze(ctx, accessRecords)
	if err != nil {
		return nil, fmt.Errorf("promotion analysis failed: %w", err)
	}

	// Step 5: Run cleanup analysis
	cleanups, err := c.cleaner.Analyze(ctx, accessRecords)
	if err != nil {
		return nil, fmt.Errorf("cleanup analysis failed: %w", err)
	}

	// Step 6: Assess memory state
	memoryState := c.assessMemoryState(contentEntries, accessRecords)

	// Step 7: Build report
	report := &CuratorReport{
		RunAt:       time.Now(),
		DryRun:      c.config.DryRun,
		Duplicates:  duplicates,
		Compactions: compactions,
		Promotions:  promotions,
		Cleanups:    cleanups,
		MemoryState: memoryState,
	}

	// Step 8: Estimate token savings
	report.TokenSavings = c.estimateSavings(report)

	return report, nil
}

// assessMemoryState evaluates the current health of the memory
func (c *CuratorService) assessMemoryState(contentEntries map[string]string, accessRecords map[string]AccessRecord) MemoryState {
	state := MemoryState{
		TotalEntries: len(contentEntries),
	}

	// Count total tokens
	for _, content := range contentEntries {
		state.TotalTokens += c.compactor.estimateTokens(content)
	}

	// Count duplicates (will be assessed on next run, use defaults here)
	state.DuplicateCount = 0
	state.OversizedCount = 0
	state.StaleCount = 0

	// Calculate health score
	healthScore := 1.0
	issues := 0

	if state.TotalEntries > 0 {
		// Deduplication penalty
		dedup, _ := c.detector.Detect(context.Background(), contentEntries)
		state.DuplicateCount = len(dedup)
		if state.DuplicateCount > 0 {
			issues += state.DuplicateCount
		}

		// Oversized entries penalty
		compactions, _ := c.compactor.Analyze(context.Background(), contentEntries)
		state.OversizedCount = len(compactions)
		if state.OversizedCount > 0 {
			issues += state.OversizedCount
		}

		// Stale entries penalty
		cleanups, _ := c.cleaner.Analyze(context.Background(), accessRecords)
		state.StaleCount = len(cleanups)
		if state.StaleCount > 0 {
			issues += state.StaleCount
		}

		// Calculate health score: reduce by 10% per issue, minimum 0
		penalty := float64(issues) * 0.1
		healthScore = 1.0 - penalty
		if healthScore < 0 {
			healthScore = 0
		}
	}

	state.HealthScore = healthScore
	return state
}

// estimateSavings calculates potential token savings from all recommendations
func (c *CuratorService) estimateSavings(report *CuratorReport) int {
	savings := 0

	// Savings from compaction
	for _, comp := range report.Compactions {
		savings += comp.SavingsTokens
	}

	// Estimate savings from duplicates (assume removing one copy)
	for _, dup := range report.Duplicates {
		if content, ok := c.getContentForID(dup.SourceID, report); ok {
			savings += c.compactor.estimateTokens(content)
		}
	}

	// Estimate savings from cleanup (all stale/low-utility content)
	for _, cleanup := range report.Cleanups {
		savings += cleanup.Size / 4 // Rough estimate: 1 token per 4 bytes
	}

	return savings
}

// getContentForID retrieves content for a given ID (helper for savings estimation)
func (c *CuratorService) getContentForID(id string, report *CuratorReport) (string, bool) {
	// This is a simplified version; in a real implementation this would
	// look up content from the actual storage
	return "", false
}

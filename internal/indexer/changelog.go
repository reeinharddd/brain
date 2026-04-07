package indexer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ChangelogEntry represents a documentation change
type ChangelogEntry struct {
	Timestamp  string   `json:"timestamp"`
	Commit     string   `json:"commit"`
	Action     string   `json:"action"` // add, modify, delete
	Domain     string   `json:"domain"`
	File       string   `json:"file"`
	Checksum   string   `json:"checksum"`
}

// ChangelogWatcher monitors docs-changelog.jsonl for changes
type ChangelogWatcher struct {
	changelogPath string
	lastPos       int64
}

// NewChangelogWatcher creates a changelog watcher
func NewChangelogWatcher(brainRoot string) *ChangelogWatcher {
	return &ChangelogWatcher{
		changelogPath: filepath.Join(brainRoot, "docs-changelog.jsonl"),
		lastPos:       0,
	}
}

// GetChangedDocs returns documents that have changed since last check
func (cw *ChangelogWatcher) GetChangedDocs() (map[string]string, error) {
	file, err := os.Open(cw.changelogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("failed to open changelog: %w", err)
	}
	defer file.Close()

	// Seek to last position
	if cw.lastPos > 0 {
		if _, err := file.Seek(cw.lastPos, 0); err != nil {
			return nil, fmt.Errorf("failed to seek changelog: %w", err)
		}
	}

	changes := make(map[string]string) // map[path]action

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry ChangelogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip malformed entries
		}

		// Track latest action for each file
		changes[entry.File] = entry.Action
	}

	// Update position for next call
	if _, err := file.Seek(0, 2); err == nil {
		cw.lastPos, _ = file.Seek(0, 2) // Seek to end
	}

	return changes, nil
}

// DeltaDetector identifies which documents need re-indexing
type DeltaDetector struct {
	brainRoot string
}

// NewDeltaDetector creates a delta detector
func NewDeltaDetector(brainRoot string) *DeltaDetector {
	return &DeltaDetector{brainRoot: brainRoot}
}

// DetectChanges returns docs that need re-indexing based on changelog
func (dd *DeltaDetector) DetectChanges(changes map[string]string) (added []string, modified []string, deleted []string) {
	for file, action := range changes {
		switch action {
		case "add":
			added = append(added, file)
		case "modify":
			modified = append(modified, file)
		case "delete":
			deleted = append(deleted, file)
		}
	}
	return
}

// ShouldPerformFullRebuild checks if incremental rebuild is appropriate
func (dd *DeltaDetector) ShouldPerformFullRebuild(changeCount int) bool {
	// If more than 30% of docs changed, do full rebuild
	return changeCount > 25 // Assume ~78 docs, 30% ~ 23
}

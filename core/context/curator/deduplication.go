package curator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
)

// DuplicateFinding represents a detected duplicate
type DuplicateFinding struct {
	SourceID    string
	TargetID    string
	Similarity  float64 // 0.0-1.0
	Method      string  // "hash" or "similarity"
	Description string
}

// DeduplicationDetector finds duplicate context entries
type DeduplicationDetector struct {
	similarityThreshold float64
}

// NewDeduplicationDetector creates a new deduplication detector
func NewDeduplicationDetector(threshold float64) *DeduplicationDetector {
	return &DeduplicationDetector{
		similarityThreshold: threshold,
	}
}

// Detect finds duplicate context entries using hash and similarity analysis
func (d *DeduplicationDetector) Detect(ctx context.Context, entries map[string]string) ([]DuplicateFinding, error) {
	if entries == nil {
		return []DuplicateFinding{}, nil
	}

	var findings []DuplicateFinding

	// Hash-based exact duplicate detection
	hashMap := make(map[string]string) // hash -> id
	for id, content := range entries {
		select {
		case <-ctx.Done():
			return findings, fmt.Errorf("deduplication cancelled: %w", ctx.Err())
		default:
		}

		hash := d.hashContent(content)
		if existingID, ok := hashMap[hash]; ok {
			findings = append(findings, DuplicateFinding{
				SourceID:    id,
				TargetID:    existingID,
				Similarity:  1.0,
				Method:      "hash",
				Description: fmt.Sprintf("Exact duplicate: %s matches %s", id, existingID),
			})
		} else {
			hashMap[hash] = id
		}
	}

	// Similarity-based detection for non-identical entries
	ids := make([]string, 0, len(entries))
	for id := range entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for i := 0; i < len(ids); i++ {
		select {
		case <-ctx.Done():
			return findings, fmt.Errorf("deduplication cancelled: %w", ctx.Err())
		default:
		}

		for j := i + 1; j < len(ids); j++ {
			// Skip if already found as hash duplicate
			if _, exists := hashMap[d.hashContent(entries[ids[i]])]; exists && hashMap[d.hashContent(entries[ids[i]])] != ids[i] {
				continue
			}
			if _, exists := hashMap[d.hashContent(entries[ids[j]])]; exists && hashMap[d.hashContent(entries[ids[j]])] != ids[j] {
				continue
			}

			sim := d.similarity(entries[ids[i]], entries[ids[j]])
			if sim >= d.similarityThreshold {
				findings = append(findings, DuplicateFinding{
					SourceID:    ids[i],
					TargetID:    ids[j],
					Similarity:  math.Round(sim*1000) / 1000,
					Method:      "similarity",
					Description: fmt.Sprintf("Similar content: %s and %s (%.3f)", ids[i], ids[j], sim),
				})
			}
		}
	}

	return findings, nil
}

// hashContent computes SHA256 hash of content
func (d *DeduplicationDetector) hashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

// similarity calculates Jaccard similarity between two strings
func (d *DeduplicationDetector) similarity(a, b string) float64 {
	wordsA := d.splitWords(a)
	wordsB := d.splitWords(b)

	if len(wordsA) == 0 && len(wordsB) == 0 {
		return 1.0
	}
	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0.0
	}

	intersection := 0
	union := make(map[string]bool)

	for w := range wordsA {
		union[w] = true
	}
	for w := range wordsB {
		if wordsA[w] {
			intersection++
		}
		union[w] = true
	}

	return float64(intersection) / float64(len(union))
}

func (d *DeduplicationDetector) splitWords(content string) map[string]bool {
	words := make(map[string]bool)
	for _, w := range strings.Fields(strings.ToLower(content)) {
		// Strip common punctuation
		w = strings.Trim(w, ".,;:!?\"'()[]{}")
		if len(w) > 0 {
			words[w] = true
		}
	}
	return words
}

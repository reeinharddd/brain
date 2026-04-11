package efficiency

import (
	"context"
	"sync"
	"time"
)

// CacheEntry represents a cached response
type CacheEntry struct {
	Content      string
	TokenCount   int
	CostUSD      float64
	CreatedAt    time.Time
	LastAccessed time.Time
	AccessCount  int
}

// ExactCache stores exact request→response mappings
type ExactCache struct {
	mu       sync.RWMutex
	items    map[string]*CacheEntry // key = hash of request
	maxItems int
	ttl      time.Duration
}

// SemanticCache stores semantically similar request mappings
type SemanticCache struct {
	mu        sync.RWMutex
	items     map[string]*CacheEntry // key = semantic hash
	maxItems  int
	threshold float64 // similarity threshold (0.0-1.0)
	ttl       time.Duration
	// requestKeys maps semantic hash to original request keys
	requestKeys map[string][]string
}

// PromptCache manages prefix/suffix caching for KV-cache optimization
type PromptCache struct {
	mu          sync.RWMutex
	prefixes    map[string]*CacheEntry // stable prefix → cache entry
	maxPrefixes int
	ttl         time.Duration
}

// NewExactCache creates a new exact cache with the given capacity and TTL
func NewExactCache(maxItems int, ttl time.Duration) *ExactCache {
	return &ExactCache{
		items:    make(map[string]*CacheEntry),
		maxItems: maxItems,
		ttl:      ttl,
	}
}

// Put stores an entry in the exact cache
func (ec *ExactCache) Put(ctx context.Context, key string, entry *CacheEntry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Evict oldest if at capacity and key doesn't exist
	if _, exists := ec.items[key]; !exists && len(ec.items) >= ec.maxItems {
		ec.evictOldestLocked()
	}

	now := time.Now()
	entry.LastAccessed = now
	entry.AccessCount = 1
	ec.items[key] = entry
	return nil
}

// Get retrieves an entry from the exact cache
func (ec *ExactCache) Get(ctx context.Context, key string) (*CacheEntry, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	default:
	}

	ec.mu.RLock()
	defer ec.mu.RUnlock()

	entry, exists := ec.items[key]
	if !exists {
		return nil, false
	}

	// Check TTL
	if time.Since(entry.CreatedAt) > ec.ttl {
		return nil, false
	}

	// Update access stats (need write lock, so upgrade)
	ec.mu.RUnlock()
	ec.mu.Lock()
	entry.LastAccessed = time.Now()
	entry.AccessCount++
	ec.mu.Unlock()
	ec.mu.RLock()

	return entry, true
}

// Delete removes an entry from the exact cache
func (ec *ExactCache) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ec.mu.Lock()
	defer ec.mu.Unlock()

	delete(ec.items, key)
	return nil
}

// Count returns the number of non-expired items in the exact cache
func (ec *ExactCache) Count() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, entry := range ec.items {
		if now.Sub(entry.CreatedAt) <= ec.ttl {
			count++
		}
	}
	return count
}

// evictOldestLocked removes the oldest entry. Must be called with write lock held.
func (ec *ExactCache) evictOldestLocked() {
	if len(ec.items) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, v := range ec.items {
		if first || v.CreatedAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.CreatedAt
			first = false
		}
	}
	delete(ec.items, oldestKey)
}

// NewSemanticCache creates a new semantic cache
func NewSemanticCache(maxItems int, threshold float64, ttl time.Duration) *SemanticCache {
	return &SemanticCache{
		items:       make(map[string]*CacheEntry),
		maxItems:    maxItems,
		threshold:   threshold,
		ttl:         ttl,
		requestKeys: make(map[string][]string),
	}
}

// Put stores an entry in the semantic cache
func (sc *SemanticCache) Put(ctx context.Context, semanticHash string, key string, entry *CacheEntry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Evict oldest if at capacity and hash doesn't exist
	if _, exists := sc.items[semanticHash]; !exists && len(sc.items) >= sc.maxItems {
		sc.evictOldestLocked()
	}

	now := time.Now()
	entry.LastAccessed = now
	entry.AccessCount = 1
	sc.items[semanticHash] = entry
	sc.requestKeys[semanticHash] = append(sc.requestKeys[semanticHash], key)
	return nil
}

// Lookup searches the semantic cache for a similar entry
func (sc *SemanticCache) Lookup(ctx context.Context, semanticHash string) (*CacheEntry, float64) {
	select {
	case <-ctx.Done():
		return nil, 0
	default:
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	entry, exists := sc.items[semanticHash]
	if !exists {
		// No exact match — try to find a close match by comparing hashes
		// For simplicity, we use a basic similarity check based on common prefix length
		var bestEntry *CacheEntry
		var bestScore float64
		now := time.Now()

		for hash, e := range sc.items {
			if now.Sub(e.CreatedAt) > sc.ttl {
				continue
			}
			score := similarityScore(semanticHash, hash)
			if score >= sc.threshold && score > bestScore {
				bestScore = score
				bestEntry = e
			}
		}
		return bestEntry, bestScore
	}

	// Check TTL
	if time.Since(entry.CreatedAt) > sc.ttl {
		return nil, 0
	}

	return entry, 1.0
}

// Get retrieves an entry from the semantic cache by its original key
func (sc *SemanticCache) Get(ctx context.Context, key string) (*CacheEntry, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	default:
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// Find the semantic hash for this key
	for hash, keys := range sc.requestKeys {
		for _, k := range keys {
			if k == key {
				entry, exists := sc.items[hash]
				if !exists {
					return nil, false
				}
				if time.Since(entry.CreatedAt) > sc.ttl {
					return nil, false
				}
				return entry, true
			}
		}
	}
	return nil, false
}

// Delete removes an entry from the semantic cache
func (sc *SemanticCache) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Find and remove from requestKeys
	for hash, keys := range sc.requestKeys {
		for i, k := range keys {
			if k == key {
				sc.requestKeys[hash] = append(keys[:i], keys[i+1:]...)
				if len(sc.requestKeys[hash]) == 0 {
					delete(sc.items, hash)
					delete(sc.requestKeys, hash)
				}
				return nil
			}
		}
	}
	return nil
}

// Count returns the number of non-expired items in the semantic cache
func (sc *SemanticCache) Count() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, entry := range sc.items {
		if now.Sub(entry.CreatedAt) <= sc.ttl {
			count++
		}
	}
	return count
}

// evictOldestLocked removes the oldest entry. Must be called with write lock held.
func (sc *SemanticCache) evictOldestLocked() {
	if len(sc.items) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, v := range sc.items {
		if first || v.CreatedAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.CreatedAt
			first = false
		}
	}
	delete(sc.items, oldestKey)
	delete(sc.requestKeys, oldestKey)
}

// NewPromptCache creates a new prompt prefix cache
func NewPromptCache(maxPrefixes int, ttl time.Duration) *PromptCache {
	return &PromptCache{
		prefixes:    make(map[string]*CacheEntry),
		maxPrefixes: maxPrefixes,
		ttl:         ttl,
	}
}

// Put stores a prefix entry in the prompt cache
func (pc *PromptCache) Put(ctx context.Context, prefix string, entry *CacheEntry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Evict oldest if at capacity and prefix doesn't exist
	if _, exists := pc.prefixes[prefix]; !exists && len(pc.prefixes) >= pc.maxPrefixes {
		pc.evictOldestLocked()
	}

	now := time.Now()
	entry.LastAccessed = now
	entry.AccessCount = 1
	pc.prefixes[prefix] = entry
	return nil
}

// Get retrieves an entry from the prompt cache by prefix match
func (pc *PromptCache) Get(ctx context.Context, prefix string) (*CacheEntry, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	default:
	}

	pc.mu.RLock()
	defer pc.mu.RUnlock()

	// Try exact prefix match first
	if entry, exists := pc.prefixes[prefix]; exists {
		if time.Since(entry.CreatedAt) <= pc.ttl {
			return entry, true
		}
		return nil, false
	}

	// Try prefix match — find the longest matching prefix
	var bestEntry *CacheEntry
	bestLen := 0
	now := time.Now()

	for storedPrefix, entry := range pc.prefixes {
		if now.Sub(entry.CreatedAt) > pc.ttl {
			continue
		}
		if len(storedPrefix) > bestLen && len(prefix) >= len(storedPrefix) {
			if prefix[:len(storedPrefix)] == storedPrefix {
				bestEntry = entry
				bestLen = len(storedPrefix)
			}
		}
	}

	return bestEntry, bestEntry != nil
}

// Delete removes an entry from the prompt cache
func (pc *PromptCache) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	delete(pc.prefixes, key)
	return nil
}

// Count returns the number of non-expired items in the prompt cache
func (pc *PromptCache) Count() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, entry := range pc.prefixes {
		if now.Sub(entry.CreatedAt) <= pc.ttl {
			count++
		}
	}
	return count
}

// evictOldestLocked removes the oldest entry. Must be called with write lock held.
func (pc *PromptCache) evictOldestLocked() {
	if len(pc.prefixes) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, v := range pc.prefixes {
		if first || v.CreatedAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.CreatedAt
			first = false
		}
	}
	delete(pc.prefixes, oldestKey)
}

// similarityScore computes a simple similarity score between two hash strings
// Returns a value between 0.0 and 1.0
func similarityScore(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Use common prefix ratio as similarity metric
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	commonPrefix := 0
	for i := 0; i < minLen; i++ {
		if a[i] == b[i] {
			commonPrefix++
		} else {
			break
		}
	}

	// Normalize by the max length
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	return float64(commonPrefix) / float64(maxLen)
}

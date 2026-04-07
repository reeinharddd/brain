package cache

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"
)

// SearchCacheKey generates cache key for search query
func SearchCacheKey(query string, limit int, domain string) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%d:%s", query, limit, domain)))
	return fmt.Sprintf("docs:search:%x", hash)
}

// StatusCacheKey returns the cache key for status
const StatusCacheKey = "docs:status"

// CacheManager wraps Cache with convenience methods
type CacheManager struct {
	cache Cache
	ttls  map[string]time.Duration
}

// NewCacheManager creates a new cache manager
func NewCacheManager(c Cache) *CacheManager {
	return &CacheManager{
		cache: c,
		ttls: map[string]time.Duration{
			"search": 1 * time.Hour,
			"status": 5 * time.Second,
		},
	}
}

// GetSearchResult retrieves cached search results
func (cm *CacheManager) GetSearchResult(ctx context.Context, query string, limit int, domain string) (interface{}, error) {
	key := SearchCacheKey(query, limit, domain)
	var result interface{}
	return result, cm.cache.Get(ctx, key, &result)
}

// SetSearchResult caches search results with 1-hour TTL
func (cm *CacheManager) SetSearchResult(ctx context.Context, query string, limit int, domain string, result interface{}) error {
	key := SearchCacheKey(query, limit, domain)
	return cm.cache.Set(ctx, key, result, cm.ttls["search"])
}

// GetStatus retrieves cached status
func (cm *CacheManager) GetStatus(ctx context.Context) (interface{}, error) {
	var result interface{}
	return result, cm.cache.Get(ctx, StatusCacheKey, &result)
}

// SetStatus caches status with 5-second TTL
func (cm *CacheManager) SetStatus(ctx context.Context, status interface{}) error {
	return cm.cache.Set(ctx, StatusCacheKey, status, cm.ttls["status"])
}

// InvalidateAll clears all caches (called on rebuild)
func (cm *CacheManager) InvalidateAll(ctx context.Context) error {
	return cm.cache.Flush(ctx)
}

// InvalidateSearches invalidates all search caches
func (cm *CacheManager) InvalidateSearches(ctx context.Context) error {
	// Redis doesn't have prefix delete, so we just flush on rebuild
	return cm.cache.Flush(ctx)
}

// Health checks cache backend health
func (cm *CacheManager) Health(ctx context.Context) error {
	return cm.cache.Health(ctx)
}

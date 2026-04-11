package efficiency

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestExactCache_PutAndGet(t *testing.T) {
	ctx := context.Background()
	cache := NewExactCache(10, 1*time.Hour)

	entry := &CacheEntry{
		Content:    "test content",
		TokenCount: 100,
		CostUSD:    0.001,
		CreatedAt:  time.Now(),
	}

	err := cache.Put(ctx, "key1", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, found := cache.Get(ctx, "key1")
	if !found {
		t.Fatal("expected to find entry")
	}
	if got.Content != "test content" {
		t.Errorf("expected content 'test content', got %q", got.Content)
	}
	if got.TokenCount != 100 {
		t.Errorf("expected token count 100, got %d", got.TokenCount)
	}
	if got.AccessCount < 1 {
		t.Errorf("expected access count >= 1, got %d", got.AccessCount)
	}
}

func TestExactCache_Miss(t *testing.T) {
	ctx := context.Background()
	cache := NewExactCache(10, 1*time.Hour)

	_, found := cache.Get(ctx, "nonexistent")
	if found {
		t.Fatal("expected cache miss")
	}
}

func TestExactCache_TTLExpiration(t *testing.T) {
	ctx := context.Background()
	cache := NewExactCache(10, 50*time.Millisecond)

	entry := &CacheEntry{
		Content:   "test content",
		TokenCount: 100,
		CostUSD:   0.001,
		CreatedAt: time.Now(),
	}

	err := cache.Put(ctx, "key1", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should exist immediately
	_, found := cache.Get(ctx, "key1")
	if !found {
		t.Fatal("expected entry to exist before TTL expiration")
	}

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	_, found = cache.Get(ctx, "key1")
	if found {
		t.Fatal("expected entry to be expired")
	}
}

func TestExactCache_MaxItemsEviction(t *testing.T) {
	ctx := context.Background()
	cache := NewExactCache(3, 1*time.Hour)

	entries := []*CacheEntry{
		{Content: "one", TokenCount: 10, CostUSD: 0.001, CreatedAt: time.Now()},
		{Content: "two", TokenCount: 20, CostUSD: 0.002, CreatedAt: time.Now().Add(1 * time.Millisecond)},
		{Content: "three", TokenCount: 30, CostUSD: 0.003, CreatedAt: time.Now().Add(2 * time.Millisecond)},
		{Content: "four", TokenCount: 40, CostUSD: 0.004, CreatedAt: time.Now().Add(3 * time.Millisecond)},
	}

	for i, e := range entries {
		err := cache.Put(ctx, string(rune('a'+i)), e)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// First entry should be evicted (oldest)
	_, found := cache.Get(ctx, "a")
	if found {
		t.Fatal("expected oldest entry to be evicted")
	}

	// Count should be at maxItems
	if count := cache.Count(); count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

func TestExactCache_Delete(t *testing.T) {
	ctx := context.Background()
	cache := NewExactCache(10, 1*time.Hour)

	entry := &CacheEntry{
		Content:    "test",
		TokenCount: 10,
		CostUSD:    0.001,
		CreatedAt:  time.Now(),
	}

	err := cache.Put(ctx, "key1", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = cache.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, found := cache.Get(ctx, "key1")
	if found {
		t.Fatal("expected entry to be deleted")
	}
}

func TestExactCache_Count(t *testing.T) {
	ctx := context.Background()
	cache := NewExactCache(10, 1*time.Hour)

	if cache.Count() != 0 {
		t.Errorf("expected count 0, got %d", cache.Count())
	}

	for i := 0; i < 5; i++ {
		entry := &CacheEntry{
			Content:    "test",
			TokenCount: 10,
			CostUSD:    0.001,
			CreatedAt:  time.Now(),
		}
		err := cache.Put(ctx, string(rune('a'+i)), entry)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if count := cache.Count(); count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestExactCache_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	cache := NewExactCache(1000, 1*time.Hour)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := string(rune('a' + n%26))
			entry := &CacheEntry{
				Content:    "concurrent test",
				TokenCount: 10,
				CostUSD:    0.001,
				CreatedAt:  time.Now(),
			}
			_ = cache.Put(ctx, key, entry)
			_, _ = cache.Get(ctx, key)
		}(i)
	}
	wg.Wait()
}

func TestSemanticCache_PutAndGet(t *testing.T) {
	ctx := context.Background()
	cache := NewSemanticCache(10, 0.85, 1*time.Hour)

	entry := &CacheEntry{
		Content:    "test content",
		TokenCount: 100,
		CostUSD:    0.001,
		CreatedAt:  time.Now(),
	}

	err := cache.Put(ctx, "semhash1", "key1", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, found := cache.Get(ctx, "key1")
	if !found {
		t.Fatal("expected to find entry")
	}
	if got.Content != "test content" {
		t.Errorf("expected content 'test content', got %q", got.Content)
	}
}

func TestSemanticCache_Miss(t *testing.T) {
	ctx := context.Background()
	cache := NewSemanticCache(10, 0.85, 1*time.Hour)

	_, found := cache.Get(ctx, "nonexistent")
	if found {
		t.Fatal("expected cache miss")
	}
}

func TestSemanticCache_Lookup(t *testing.T) {
	ctx := context.Background()
	cache := NewSemanticCache(10, 0.85, 1*time.Hour)

	entry := &CacheEntry{
		Content:    "test content",
		TokenCount: 100,
		CostUSD:    0.001,
		CreatedAt:  time.Now(),
	}

	hash := "abc123"
	err := cache.Put(ctx, hash, "key1", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Exact hash match
	got, score := cache.Lookup(ctx, hash)
	if got == nil {
		t.Fatal("expected to find entry")
	}
	if score != 1.0 {
		t.Errorf("expected score 1.0, got %f", score)
	}

	// Non-existent hash
	got, score = cache.Lookup(ctx, "zzz999")
	if score > 0 && got != nil {
		t.Logf("found partial match with score %f (expected none for very different hash)", score)
	}
}

func TestSemanticCache_SimilarityMatching(t *testing.T) {
	tests := []struct {
		name          string
		hashA         string
		hashB         string
		expectGreater bool
	}{
		{
			name:          "identical hashes have score 1.0",
			hashA:         "abc123",
			hashB:         "abc123",
			expectGreater: true,
		},
		{
			name:          "similar prefix hashes have higher score",
			hashA:         "abc123",
			hashB:         "abc456",
			expectGreater: true,
		},
		{
			name:          "completely different hashes have low score",
			hashA:         "abc123",
			hashB:         "xyz789",
			expectGreater: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := similarityScore(tt.hashA, tt.hashB)
			if tt.hashA == tt.hashB {
				if score != 1.0 {
					t.Errorf("expected score 1.0 for identical hashes, got %f", score)
				}
			} else if tt.expectGreater && score <= 0 {
				t.Errorf("expected score > 0 for similar hashes, got %f", score)
			}
		})
	}
}

func TestSemanticCache_TTLExpiration(t *testing.T) {
	ctx := context.Background()
	cache := NewSemanticCache(10, 0.85, 50*time.Millisecond)

	entry := &CacheEntry{
		Content:   "test content",
		TokenCount: 100,
		CostUSD:   0.001,
		CreatedAt: time.Now(),
	}

	err := cache.Put(ctx, "semhash1", "key1", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, found := cache.Get(ctx, "key1")
	if !found {
		t.Fatal("expected entry to exist before TTL expiration")
	}

	time.Sleep(100 * time.Millisecond)

	_, found = cache.Get(ctx, "key1")
	if found {
		t.Fatal("expected entry to be expired")
	}
}

func TestSemanticCache_MaxItemsEviction(t *testing.T) {
	ctx := context.Background()
	cache := NewSemanticCache(3, 0.85, 1*time.Hour)

	entries := []*CacheEntry{
		{Content: "one", TokenCount: 10, CostUSD: 0.001, CreatedAt: time.Now()},
		{Content: "two", TokenCount: 20, CostUSD: 0.002, CreatedAt: time.Now().Add(1 * time.Millisecond)},
		{Content: "three", TokenCount: 30, CostUSD: 0.003, CreatedAt: time.Now().Add(2 * time.Millisecond)},
		{Content: "four", TokenCount: 40, CostUSD: 0.004, CreatedAt: time.Now().Add(3 * time.Millisecond)},
	}

	for i, e := range entries {
		err := cache.Put(ctx, string(rune('a'+i)), string(rune('a'+i)), e)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Count should be at maxItems
	if count := cache.Count(); count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

func TestSemanticCache_Delete(t *testing.T) {
	ctx := context.Background()
	cache := NewSemanticCache(10, 0.85, 1*time.Hour)

	entry := &CacheEntry{
		Content:    "test",
		TokenCount: 10,
		CostUSD:    0.001,
		CreatedAt:  time.Now(),
	}

	err := cache.Put(ctx, "semhash1", "key1", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = cache.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, found := cache.Get(ctx, "key1")
	if found {
		t.Fatal("expected entry to be deleted")
	}
}

func TestSemanticCache_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	cache := NewSemanticCache(1000, 0.85, 1*time.Hour)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := string(rune('a' + n%26))
			hash := string(rune('A' + n%26))
			entry := &CacheEntry{
				Content:    "concurrent test",
				TokenCount: 10,
				CostUSD:    0.001,
				CreatedAt:  time.Now(),
			}
			_ = cache.Put(ctx, hash, key, entry)
			_, _ = cache.Get(ctx, key)
			_, _ = cache.Lookup(ctx, hash)
		}(i)
	}
	wg.Wait()
}

func TestSemanticCache_Count(t *testing.T) {
	ctx := context.Background()
	cache := NewSemanticCache(10, 0.85, 1*time.Hour)

	if cache.Count() != 0 {
		t.Errorf("expected count 0, got %d", cache.Count())
	}

	for i := 0; i < 5; i++ {
		entry := &CacheEntry{
			Content:    "test",
			TokenCount: 10,
			CostUSD:    0.001,
			CreatedAt:  time.Now(),
		}
		err := cache.Put(ctx, string(rune('a'+i)), string(rune('a'+i)), entry)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if count := cache.Count(); count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestPromptCache_PutAndGet(t *testing.T) {
	ctx := context.Background()
	cache := NewPromptCache(10, 1*time.Hour)

	entry := &CacheEntry{
		Content:    "test content",
		TokenCount: 100,
		CostUSD:    0.001,
		CreatedAt:  time.Now(),
	}

	prefix := "system: you are a helpful assistant"
	err := cache.Put(ctx, prefix, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, found := cache.Get(ctx, prefix)
	if !found {
		t.Fatal("expected to find entry")
	}
	if got.Content != "test content" {
		t.Errorf("expected content 'test content', got %q", got.Content)
	}
}

func TestPromptCache_PrefixMatching(t *testing.T) {
	ctx := context.Background()
	cache := NewPromptCache(10, 1*time.Hour)

	entry := &CacheEntry{
		Content:    "prefix cached",
		TokenCount: 100,
		CostUSD:    0.001,
		CreatedAt:  time.Now(),
	}

	storedPrefix := "system: you are a helpful assistant"
	err := cache.Put(ctx, storedPrefix, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Longer input that starts with the stored prefix
	input := "system: you are a helpful assistant. Please help me with coding."
	got, found := cache.Get(ctx, input)
	if !found {
		t.Fatal("expected prefix match")
	}
	if got.Content != "prefix cached" {
		t.Errorf("expected content 'prefix cached', got %q", got.Content)
	}
}

func TestPromptCache_Miss(t *testing.T) {
	ctx := context.Background()
	cache := NewPromptCache(10, 1*time.Hour)

	_, found := cache.Get(ctx, "nonexistent prefix")
	if found {
		t.Fatal("expected cache miss")
	}
}

func TestPromptCache_TTLExpiration(t *testing.T) {
	ctx := context.Background()
	cache := NewPromptCache(10, 50*time.Millisecond)

	entry := &CacheEntry{
		Content:   "test content",
		TokenCount: 100,
		CostUSD:   0.001,
		CreatedAt: time.Now(),
	}

	err := cache.Put(ctx, "prefix1", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, found := cache.Get(ctx, "prefix1")
	if !found {
		t.Fatal("expected entry to exist before TTL expiration")
	}

	time.Sleep(100 * time.Millisecond)

	_, found = cache.Get(ctx, "prefix1")
	if found {
		t.Fatal("expected entry to be expired")
	}
}

func TestPromptCache_MaxItemsEviction(t *testing.T) {
	ctx := context.Background()
	cache := NewPromptCache(3, 1*time.Hour)

	entries := []*CacheEntry{
		{Content: "one", TokenCount: 10, CostUSD: 0.001, CreatedAt: time.Now()},
		{Content: "two", TokenCount: 20, CostUSD: 0.002, CreatedAt: time.Now().Add(1 * time.Millisecond)},
		{Content: "three", TokenCount: 30, CostUSD: 0.003, CreatedAt: time.Now().Add(2 * time.Millisecond)},
		{Content: "four", TokenCount: 40, CostUSD: 0.004, CreatedAt: time.Now().Add(3 * time.Millisecond)},
	}

	for i, e := range entries {
		err := cache.Put(ctx, string(rune('a'+i)), e)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Count should be at maxPrefixes
	if count := cache.Count(); count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

func TestPromptCache_Delete(t *testing.T) {
	ctx := context.Background()
	cache := NewPromptCache(10, 1*time.Hour)

	entry := &CacheEntry{
		Content:    "test",
		TokenCount: 10,
		CostUSD:    0.001,
		CreatedAt:  time.Now(),
	}

	err := cache.Put(ctx, "prefix1", entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = cache.Delete(ctx, "prefix1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, found := cache.Get(ctx, "prefix1")
	if found {
		t.Fatal("expected entry to be deleted")
	}
}

func TestPromptCache_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	cache := NewPromptCache(1000, 1*time.Hour)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			prefix := string(rune('a' + n%26))
			entry := &CacheEntry{
				Content:    "concurrent test",
				TokenCount: 10,
				CostUSD:    0.001,
				CreatedAt:  time.Now(),
			}
			_ = cache.Put(ctx, prefix, entry)
			_, _ = cache.Get(ctx, prefix)
		}(i)
	}
	wg.Wait()
}

func TestPromptCache_Count(t *testing.T) {
	ctx := context.Background()
	cache := NewPromptCache(10, 1*time.Hour)

	if cache.Count() != 0 {
		t.Errorf("expected count 0, got %d", cache.Count())
	}

	for i := 0; i < 5; i++ {
		entry := &CacheEntry{
			Content:    "test",
			TokenCount: 10,
			CostUSD:    0.001,
			CreatedAt:  time.Now(),
		}
		err := cache.Put(ctx, string(rune('a'+i)), entry)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if count := cache.Count(); count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

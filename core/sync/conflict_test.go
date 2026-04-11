package sync

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"
)

func makeConflict(local, remote string, localVer, remoteVer string) Conflict {
	return Conflict{
		LocalData:     []byte(local),
		RemoteData:    []byte(remote),
		ConflictType:  "content",
		LocalVersion:  localVer,
		RemoteVersion: remoteVer,
		DetectedAt:    time.Now(),
	}
}

func TestConflictResolver_LastWriteWins(t *testing.T) {
	resolver := NewConflictResolver(LastWriteWins)
	ctx := context.Background()

	tests := []struct {
		name         string
		conflict     Conflict
		wantWinner   string
	}{
		{
			name:       "local version is newer",
			conflict:   makeConflict("local data", "remote data", "v2", "v1"),
			wantWinner: "local",
		},
		{
			name:       "remote version is newer",
			conflict:   makeConflict("local data", "remote data", "v1", "v2"),
			wantWinner: "remote",
		},
		{
			name:       "same version defaults to remote",
			conflict:   makeConflict("local data", "remote data", "v1", "v1"),
			wantWinner: "remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolution, err := resolver.Resolve(ctx, tt.conflict, "default")
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}

			if resolution.Winner != tt.wantWinner {
				t.Errorf("Winner = %s, want %s", resolution.Winner, tt.wantWinner)
			}

			if resolution.RequiresApproval {
				t.Error("last_write_wins should not require approval")
			}

			if tt.wantWinner == "local" && !bytes.Equal(resolution.MergedData, tt.conflict.LocalData) {
				t.Error("merged data should be local data")
			}
			if tt.wantWinner == "remote" && !bytes.Equal(resolution.MergedData, tt.conflict.RemoteData) {
				t.Error("merged data should be remote data")
			}
		})
	}
}

func TestConflictResolver_LocalPreferred(t *testing.T) {
	resolver := NewConflictResolver(LocalPreferred)
	ctx := context.Background()

	conflict := makeConflict("local data", "remote data", "v1", "v2")

	resolution, err := resolver.Resolve(ctx, conflict, "default")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if resolution.Winner != "local" {
		t.Errorf("Winner = %s, want local", resolution.Winner)
	}

	if !bytes.Equal(resolution.MergedData, conflict.LocalData) {
		t.Error("merged data should be local data")
	}

	if resolution.RequiresApproval {
		t.Error("local_preferred should not require approval")
	}
}

func TestConflictResolver_RemotePreferred(t *testing.T) {
	resolver := NewConflictResolver(RemotePreferred)
	ctx := context.Background()

	conflict := makeConflict("local data", "remote data", "v2", "v1")

	resolution, err := resolver.Resolve(ctx, conflict, "default")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if resolution.Winner != "remote" {
		t.Errorf("Winner = %s, want remote", resolution.Winner)
	}

	if !bytes.Equal(resolution.MergedData, conflict.RemoteData) {
		t.Error("merged data should be remote data")
	}

	if resolution.RequiresApproval {
		t.Error("remote_preferred should not require approval")
	}
}

func TestConflictResolver_Merge(t *testing.T) {
	resolver := NewConflictResolver(Merge)
	ctx := context.Background()

	conflict := makeConflict("local data", "remote data", "v1", "v1")

	resolution, err := resolver.Resolve(ctx, conflict, "default")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if resolution.Winner != "merged" {
		t.Errorf("Winner = %s, want merged", resolution.Winner)
	}

	expectedLocal := []byte("local data")
	if !bytes.Contains(resolution.MergedData, expectedLocal) {
		t.Error("merged data should contain local data")
	}

	expectedRemote := []byte("remote data")
	if !bytes.Contains(resolution.MergedData, expectedRemote) {
		t.Error("merged data should contain remote data")
	}

	if resolution.RequiresApproval {
		t.Error("merge should not require approval")
	}
}

func TestConflictResolver_Manual(t *testing.T) {
	resolver := NewConflictResolver(Manual)
	ctx := context.Background()

	conflict := makeConflict("local data", "remote data", "v1", "v1")

	resolution, err := resolver.Resolve(ctx, conflict, "default")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if resolution.Winner != "pending" {
		t.Errorf("Winner = %s, want pending", resolution.Winner)
	}

	if resolution.MergedData != nil {
		t.Error("manual resolution should have nil merged data")
	}

	if !resolution.RequiresApproval {
		t.Error("manual should require approval")
	}
}

func TestConflictResolver_TypeSpecificStrategy(t *testing.T) {
	resolver := NewConflictResolver(LastWriteWins)
	resolver.SetTypeStrategy("notes", Manual)
	resolver.SetTypeStrategy("memories", Merge)

	ctx := context.Background()
	conflict := makeConflict("data", "other", "v1", "v2")

	tests := []struct {
		name       string
		artType    string
		wantWinner string
	}{
		{name: "notes uses manual", artType: "notes", wantWinner: "pending"},
		{name: "memories uses merge", artType: "memories", wantWinner: "merged"},
		{name: "default uses last_write_wins", artType: "default", wantWinner: "remote"},
		{name: "unknown type uses default", artType: "unknown", wantWinner: "remote"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolution, err := resolver.Resolve(ctx, conflict, tt.artType)
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}

			if resolution.Winner != tt.wantWinner {
				t.Errorf("Winner = %s, want %s", resolution.Winner, tt.wantWinner)
			}
		})
	}
}

func TestConflictResolver_ContextCancellation(t *testing.T) {
	resolver := NewConflictResolver(LastWriteWins)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	conflict := makeConflict("data", "other", "v1", "v2")

	_, err := resolver.Resolve(ctx, conflict, "default")
	if err == nil {
		t.Error("Resolve() with cancelled context should fail")
	}
}

func TestConflictResolver_ConcurrentResolution(t *testing.T) {
	resolver := NewConflictResolver(LastWriteWins)
	ctx := context.Background()

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conflict := makeConflict("local", "remote", "v1", "v2")
			_, err := resolver.Resolve(ctx, conflict, "default")
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent resolution error: %v", err)
	}
}

func TestConflictResolver_UnknownStrategy(t *testing.T) {
	resolver := NewConflictResolver(ConflictResolutionStrategy("unknown"))
	ctx := context.Background()

	conflict := makeConflict("data", "other", "v1", "v2")

	_, err := resolver.Resolve(ctx, conflict, "default")
	if err == nil {
		t.Error("Resolve() with unknown strategy should fail")
	}
}

func TestConflictResolver_SetTypeStrategyConcurrent(t *testing.T) {
	resolver := NewConflictResolver(LastWriteWins)

	var wg sync.WaitGroup

	// Concurrent writes to type strategies
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			resolver.SetTypeStrategy("type_"+string(rune('A'+id)), LocalPreferred)
		}(i)
	}

	wg.Wait()

	// Verify no data race and strategies are set
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		conflict := makeConflict("local", "remote", "v1", "v2")
		resolution, err := resolver.Resolve(ctx, conflict, "type_"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("Resolve() error for type_%c: %v", 'A'+i, err)
		}
		if resolution.Winner != "local" {
			t.Errorf("type_%c winner = %s, want local", 'A'+i, resolution.Winner)
		}
	}
}

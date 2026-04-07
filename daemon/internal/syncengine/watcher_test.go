package syncengine

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileWatcherTriggersSyncOnFileChange(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "registry.yml")

	if err := os.WriteFile(sourcePath, []byte("initial: true\n"), 0644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	triggered := make(chan struct{}, 1)
	logCh := make(chan string, 10)
	watcher, err := NewFileWatcher(logCh, func() error {
		select {
		case triggered <- struct{}{}:
		default:
		}
		return nil
	})
	if err != nil {
		t.Fatalf("new watcher: %v", err)
	}
	defer watcher.Stop()

	if err := watcher.Start(sourcePath); err != nil {
		t.Fatalf("start watcher: %v", err)
	}

	if err := os.WriteFile(sourcePath, []byte("initial: false\n"), 0644); err != nil {
		t.Fatalf("modify source file: %v", err)
	}

	select {
	case <-triggered:
	case <-time.After(3 * time.Second):
		t.Fatal("expected sync callback to fire after file change")
	}
}

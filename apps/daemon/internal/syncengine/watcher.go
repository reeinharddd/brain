package syncengine

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher  *fsnotify.Watcher
	logger   chan<- string
	syncFn   func() error
	debounce time.Duration
	mu       sync.Mutex
	pending  bool
	stopCh   chan struct{}
	once     sync.Once
}

func NewFileWatcher(logger chan<- string, syncFn func() error) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:  watcher,
		logger:   logger,
		syncFn:   syncFn,
		debounce: 250 * time.Millisecond,
		stopCh:   make(chan struct{}),
	}, nil
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func (f *FileWatcher) Start(paths ...string) error {
	for _, path := range paths {
		if err := f.addPath(path); err != nil {
			return err
		}
	}

	f.logger <- fmt.Sprintf("[SyncWatcher] Watching %d source paths", len(paths))
	go f.loop()
	return nil
}

func (f *FileWatcher) Stop() error {
	var err error
	f.once.Do(func() {
		close(f.stopCh)
		err = f.watcher.Close()
	})
	return err
}

func (f *FileWatcher) addPath(path string) error {
	expanded := expandHome(path)
	info, err := os.Stat(expanded)
	if err != nil {
		return fmt.Errorf("sync watcher cannot stat %s: %w", expanded, err)
	}

	watchPath := expanded
	if !info.IsDir() {
		watchPath = filepath.Dir(expanded)
	}

	if err := f.watcher.Add(watchPath); err != nil {
		return fmt.Errorf("sync watcher cannot watch %s: %w", watchPath, err)
	}

	f.logger <- fmt.Sprintf("[SyncWatcher] Watching %s", watchPath)
	return nil
}

func (f *FileWatcher) loop() {
	for {
		select {
		case event, ok := <-f.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				f.logger <- fmt.Sprintf("[SyncWatcher] Change detected: %s (%s)", event.Name, event.Op.String())
				f.scheduleSync()
			}
		case err, ok := <-f.watcher.Errors:
			if !ok {
				return
			}
			f.logger <- fmt.Sprintf("[SyncWatcher] Error: %v", err)
		case <-f.stopCh:
			return
		}
	}
}

func (f *FileWatcher) scheduleSync() {
	f.mu.Lock()
	if f.pending {
		f.mu.Unlock()
		return
	}
	f.pending = true
	f.mu.Unlock()

	time.AfterFunc(f.debounce, func() {
		f.mu.Lock()
		f.pending = false
		f.mu.Unlock()

		if err := f.syncFn(); err != nil {
			f.logger <- fmt.Sprintf("[SyncWatcher] Sync failed: %v", err)
			return
		}
		f.logger <- "[SyncWatcher] Sync completed"
	})
}

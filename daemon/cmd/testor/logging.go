package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogRecord struct {
	TS         string `json:"ts"`
	Event      string `json:"event"`
	Suite      string `json:"suite,omitempty"`
	Name       string `json:"name,omitempty"`
	File       string `json:"file,omitempty"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
	TotalTests int    `json:"total_tests,omitempty"`
	Passed     int    `json:"passed,omitempty"`
	Failed     int    `json:"failed,omitempty"`
	Skipped    int    `json:"skipped,omitempty"`
	DurationMS int64  `json:"duration_ms,omitempty"`
	DurationS  int64  `json:"duration_sec,omitempty"`
}

type Logger struct {
	mu   sync.Mutex
	out  io.Writer
	file *os.File
}

func NewLogger(root, outputDir string) (*Logger, error) {
	if outputDir == "" {
		outputDir = ".logs"
	}

	fullDir := filepath.Join(root, outputDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return &Logger{out: os.Stdout}, nil
	}

	name := fmt.Sprintf("test-run-%s.ndjson", time.Now().UTC().Format("2006-01-02-15-04-05"))
	logPath := filepath.Join(fullDir, name)
	f, err := os.Create(logPath)
	if err != nil {
		return &Logger{out: os.Stdout}, nil
	}

	return &Logger{out: os.Stdout, file: f}, nil
}

func (l *Logger) Event(record LogRecord) {
	record.TS = time.Now().UTC().Format(time.RFC3339)

	line, err := json.Marshal(record)
	if err != nil {
		return
	}
	line = append(line, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()

	_, _ = l.out.Write(line)
	if l.file != nil {
		_, _ = l.file.Write(line)
	}
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}

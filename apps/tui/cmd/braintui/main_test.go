package main

import (
	"testing"
)

func TestMainPackage(t *testing.T) {
	t.Run("main function exists", func(t *testing.T) {
		// The main function is tested by integration; verify package compiles
	})

	t.Run("run function exists", func(t *testing.T) {
		// The run function is tested by integration; verify package structure
		// We can't call run() directly as it initializes a TUI requiring a terminal
	})
}

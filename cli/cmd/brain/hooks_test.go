package main

import "testing"

func TestBlockEnvWritesResultBlocksEnvFilePaths(t *testing.T) {
	message, blocked := blockEnvWritesResult("write file to .env.local")
	if !blocked {
		t.Fatal("expected .env path to be blocked")
	}
	if message == "" {
		t.Fatal("expected block message")
	}
}

func TestBlockEnvWritesResultBlocksPrivateKeys(t *testing.T) {
	message, blocked := blockEnvWritesResult("-----BEGIN OPENSSH PRIVATE KEY-----")
	if !blocked {
		t.Fatal("expected private key to be blocked")
	}
	if message == "" {
		t.Fatal("expected block message")
	}
}

func TestLinterForFileUsesExpectedToolSelection(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		expected string
	}{
		{name: "typescript", ext: "ts", expected: "Biome"},
		{name: "python", ext: "py", expected: "Ruff"},
		{name: "shell", ext: "sh", expected: "ShellCheck"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			label, _, _, _ := linterSelectionForExt(tc.ext)
			if label != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, label)
			}
		})
	}
}

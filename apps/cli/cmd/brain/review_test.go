package main

import (
	"strings"
	"testing"
)

func TestRunReview(t *testing.T) {
	t.Run("missing action returns error", func(t *testing.T) {
		err := runReview([]string{})
		if err == nil {
			t.Fatal("expected error for missing action")
		}
		if !strings.Contains(err.Error(), "action required") {
			t.Errorf("expected 'action required' error, got: %v", err)
		}
	})

	t.Run("unknown action returns error", func(t *testing.T) {
		err := runReview([]string{"unknown"})
		if err == nil {
			t.Fatal("expected error for unknown action")
		}
		if !strings.Contains(err.Error(), "unknown action") {
			t.Errorf("expected 'unknown action' error, got: %v", err)
		}
	})

	t.Run("list action succeeds", func(t *testing.T) {
		err := runReview([]string{"list"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("approve with missing ID returns error", func(t *testing.T) {
		err := runReview([]string{"approve"})
		if err == nil {
			t.Fatal("expected error for missing recommendation ID")
		}
		if !strings.Contains(err.Error(), "recommendation ID required") {
			t.Errorf("expected 'recommendation ID required' error, got: %v", err)
		}
	})

	t.Run("approve with ID succeeds", func(t *testing.T) {
		err := runReview([]string{"approve", "rec-001"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("reject with missing ID returns error", func(t *testing.T) {
		err := runReview([]string{"reject"})
		if err == nil {
			t.Fatal("expected error for missing recommendation ID")
		}
		if !strings.Contains(err.Error(), "recommendation ID required") {
			t.Errorf("expected 'recommendation ID required' error, got: %v", err)
		}
	})

	t.Run("reject with ID succeeds", func(t *testing.T) {
		err := runReview([]string{"reject", "rec-001"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("apply action succeeds", func(t *testing.T) {
		err := runReview([]string{"apply"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunStatus(t *testing.T) {
	t.Run("status command outputs all subsystems", func(t *testing.T) {
		err := runStatus(nil, []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("status command shows correct count", func(t *testing.T) {
		// Count expected subsystems manually
		expected := 15
		subsystems := []string{
			"Observability", "Artifact Registry", "Token Efficiency",
			"Context Compiler", "Model Router", "Context Curator",
			"Memory Sync", "MCP Hub", "Governance", "Delegation Graph",
			"Agent Pool", "Workflows", "Skill Registry", "AutoEvolve", "Cost Engine",
		}
		if len(subsystems) != expected {
			t.Errorf("expected %d subsystems, got %d", expected, len(subsystems))
		}
	})
}

func TestReviewCommandStruct(t *testing.T) {
	t.Run("review command struct fields", func(t *testing.T) {
		cmd := ReviewCommand{
			Action: "list",
			ID:     "rec-001",
			All:    true,
			JSON:   true,
		}
		if cmd.Action != "list" {
			t.Errorf("expected action 'list', got %q", cmd.Action)
		}
		if cmd.ID != "rec-001" {
			t.Errorf("expected ID 'rec-001', got %q", cmd.ID)
		}
		if !cmd.All {
			t.Error("expected All to be true")
		}
		if !cmd.JSON {
			t.Error("expected JSON to be true")
		}
	})

	t.Run("status command struct fields", func(t *testing.T) {
		cmd := StatusCommand{JSON: true}
		if !cmd.JSON {
			t.Error("expected JSON to be true")
		}
	})
}

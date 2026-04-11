package app

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestAppNew(t *testing.T) {
	t.Run("creates a valid App", func(t *testing.T) {
		a, err := New()
		if err != nil {
			t.Fatalf("New() returned error: %v", err)
		}
		if a == nil {
			t.Fatal("New() returned nil")
		}
	})

	t.Run("all expected fields initialized", func(t *testing.T) {
		a, err := New()
		if err != nil {
			t.Fatalf("New() returned error: %v", err)
		}
		if a.app == nil {
			t.Error("app field is nil")
		}
		if a.pages == nil {
			t.Error("pages field is nil")
		}
		if a.status == nil {
			t.Error("status field is nil")
		}
		if a.header == nil {
			t.Error("header field is nil")
		}
		if a.footer == nil {
			t.Error("footer field is nil")
		}
		if a.daemonURL == "" {
			t.Error("daemonURL is empty")
		}
	})
}

func TestAppShowStatus(t *testing.T) {
	t.Run("produces non-empty output", func(t *testing.T) {
		a := newTestApp()
		a.showStatus()
		text := a.status.GetText(false)
		if text == "" {
			t.Fatal("showStatus produced empty output")
		}
	})

	t.Run("all subsystem names appear", func(t *testing.T) {
		a := newTestApp()
		a.showStatus()
		text := a.status.GetText(false)

		expectedSubsystems := []string{
			"Observability",
			"Artifact Registry",
			"Token Efficiency",
			"Context Compiler",
			"Model Router",
			"Context Curator",
			"Memory Sync",
			"MCP Hub",
			"Governance",
			"Delegation Graph",
			"Agent Pool",
			"Workflows",
			"Skill Registry",
			"AutoEvolve",
			"Cost Engine",
		}

		for _, name := range expectedSubsystems {
			if !strings.Contains(text, name) {
				t.Errorf("status output missing subsystem: %s", name)
			}
		}
	})

	t.Run("subsystem count is 15", func(t *testing.T) {
		a := newTestApp()
		a.showStatus()
		text := a.status.GetText(false)
		if !strings.Contains(text, "Total: 15 subsystems operational") {
			t.Errorf("expected 15 subsystems, got output without count")
		}
	})
}

func TestAppShowArtifacts(t *testing.T) {
	t.Run("produces non-empty output", func(t *testing.T) {
		a := newTestApp()
		a.showArtifacts()
		text := a.status.GetText(false)
		if text == "" {
			t.Fatal("showArtifacts produced empty output")
		}
	})
}

func TestAppShowContext(t *testing.T) {
	t.Run("produces non-empty output", func(t *testing.T) {
		a := newTestApp()
		a.showContext()
		text := a.status.GetText(false)
		if text == "" {
			t.Fatal("showContext produced empty output")
		}
	})
}

func TestAppShowPolicy(t *testing.T) {
	t.Run("produces non-empty output", func(t *testing.T) {
		a := newTestApp()
		a.showPolicy()
		text := a.status.GetText(false)
		if text == "" {
			t.Fatal("showPolicy produced empty output")
		}
	})
}

func TestAppShowEvents(t *testing.T) {
	t.Run("produces non-empty output", func(t *testing.T) {
		a := newTestApp()
		a.showEvents()
		text := a.status.GetText(false)
		if text == "" {
			t.Fatal("showEvents produced empty output")
		}
	})
}

func newTestApp() *App {
	header := tview.NewTextView()
	header.SetTextColor(tcell.ColorBlack)
	header.SetBackgroundColor(tcell.ColorDarkCyan)
	header.SetDynamicColors(true)

	footer := tview.NewTextView()
	footer.SetTextColor(tcell.ColorWhite)
	footer.SetBackgroundColor(tcell.ColorBlack)

	status := tview.NewTextView()
	status.SetDynamicColors(true)
	status.SetScrollable(true)
	status.SetWordWrap(true)

	return &App{
		app:       tview.NewApplication(),
		pages:     tview.NewPages(),
		status:    status,
		header:    header,
		footer:    footer,
		daemonURL: "http://localhost:8080",
	}
}

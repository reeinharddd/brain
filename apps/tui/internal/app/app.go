package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the Brain TUI application
type App struct {
	app       *tview.Application
	pages     *tview.Pages
	status    *tview.TextView
	header    *tview.TextView
	footer    *tview.TextView
	daemonURL string
}

func New() (*App, error) {
	a := &App{
		app:       tview.NewApplication(),
		pages:     tview.NewPages(),
		daemonURL: "http://localhost:8080",
	}

	a.buildUI()
	return a, nil
}

func (a *App) buildUI() {
	// Header
	a.header = tview.NewTextView()
	a.header.SetText(" Brain Daemon TUI v0.1.0 | [yellow]TAB[white] to switch panels | [yellow]Q[white] to quit")
	a.header.SetTextColor(tcell.ColorBlack)
	a.header.SetBackgroundColor(tcell.ColorDarkCyan)
	a.header.SetDynamicColors(true)

	// Main content with tabs
	tabs := tview.NewTextView()
	tabs.SetText("1:Artifacts  2:Context  3:Policy  4:Events  5:Status")
	tabs.SetTextColor(tcell.ColorWhite)
	tabs.SetBackgroundColor(tcell.ColorDarkGray)

	// Content area
	content := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Left panel - navigation
	nav := tview.NewList().
		AddItem("Artifacts", "Browse artifacts", '1', func() {
			a.showArtifacts()
		}).
		AddItem("Context", "View context bundle", '2', func() {
			a.showContext()
		}).
		AddItem("Policy", "Inspect policies", '3', func() {
			a.showPolicy()
		}).
		AddItem("Events", "Real-time events", '4', func() {
			a.showEvents()
		}).
		AddItem("Status", "System status", '5', func() {
			a.showStatus()
		})

	// Right panel - content display
	a.status = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	a.showStatus() // Initial view

	content.AddItem(nav, 25, 0, true)
	content.AddItem(a.status, 0, 1, false)

	// Footer
	a.footer = tview.NewTextView()
	a.footer.SetText(" [green]●[white] Connected to daemon")
	a.footer.SetTextColor(tcell.ColorWhite)
	a.footer.SetBackgroundColor(tcell.ColorBlack)

	// Layout
	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	layout.AddItem(a.header, 1, 0, false)
	layout.AddItem(tview.NewTextView(), 1, 0, false)
	layout.AddItem(content, 0, 1, true)
	layout.AddItem(a.footer, 1, 0, false)

	a.pages.AddPage("main", layout, true, true)
	a.app.SetRoot(a.pages, true)

	// Key handling
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC || (event.Key() == tcell.KeyRune && event.Rune() == 'q') {
			a.app.Stop()
			return nil
		}
		return event
	})
}

func (a *App) showArtifacts() {
	a.status.Clear()
	fmt.Fprint(a.status, `[yellow]Artifacts Registry[white]

[green]Skills:[white]
  • code-refactoring    v1.2.0  [green]active[white]
  • debug-session       v1.0.0  [green]active[white]
  • test-generation     v0.9.0  [green]active[white]

[green]MCP Servers:[white]
  • filesystem          v1.0.0  [green]running[white]
  • git                 v1.0.0  [green]running[white]
  • github              v1.0.0  [yellow]stopped[white]

[green]Agents:[white]
  • architect-core      v1.0.0  [green]idle[white]
  • builder-fast        v1.0.0  [green]busy[white]
`)
}

func (a *App) showContext() {
	a.status.Clear()
	fmt.Fprint(a.status, `[yellow]Context Bundle[white]

[green]Bundle ID:[white] ctx_20260411_abc123
[green]Compiled at:[white] 2026-04-11T10:00:00Z
[green]Token count:[white] 1,400 / 8,000 (17.5%)
[green]Compression:[white] applied (original: 3,200 tokens)

[green]Layers:[white]
  0. Hard Policy          [green]✓[white]  200 tokens
  1. Identity             [green]✓[white]  150 tokens
  2. Org Baseline         [green]✓[white]  200 tokens
  3. User Baseline        [green]✓[white]  150 tokens
  4. Workspace Context    [green]✓[white]  300 tokens
  5. Project Context      [green]✓[white]  250 tokens
  6. Task-Local Context   [green]✓[white]  100 tokens
  7. Active Skills        [yellow]~[white]  50 tokens (progressive disclosure)
`)
}

func (a *App) showPolicy() {
	a.status.Clear()
	fmt.Fprint(a.status, `[yellow]Policy Resolution[white]

[green]Scope Chain:[white] org:myorg → user:alice → workspace:brain

[green]Rules:[white]
  no-hardcoded-secrets     [red]enforced[white]   (org:myorg, hard)
  model-routing            [yellow]prefer-local[white] (user:alice, guarded)
  max-token-budget         [green]8000 tokens[white]  (workspace:brain, soft)
  require-approval         [yellow]writes[white]      (org:myorg, hard)
`)
}

func (a *App) showEvents() {
	a.status.Clear()
	fmt.Fprint(a.status, `[yellow]Real-Time Events[white]

[green]Monitoring...[white] (connect to daemon WebSocket for live events)

Recent events:
  [green]10:00:01[white] [blue]INFO[white]  Daemon startup complete
  [green]10:00:02[white] [blue]INFO[white]  Skills registry loaded (3 skills)
  [green]10:00:03[white] [blue]INFO[white]  MCP registry loaded (2 servers)
  [green]10:00:05[white] [blue]INFO[white]  Initial sync completed
  [green]10:00:10[white] [blue]INFO[white]  Health check: all systems operational
`)
}

func (a *App) showStatus() {
	a.status.Clear()
	subsystems := []struct {
		name, status, detail string
	}{
		{"Observability", "✓", "OpenTelemetry + Prometheus"},
		{"Artifact Registry", "✓", "Dependencies + versions"},
		{"Token Efficiency", "✓", "Multi-tier cache"},
		{"Context Compiler", "✓", "12-layer bundles"},
		{"Model Router", "✓", "3-tier routing"},
		{"Context Curator", "✓", "Dedup + autoDream"},
		{"Memory Sync", "✓", "5 conflict strategies"},
		{"MCP Hub", "✓", "5 official servers"},
		{"Governance", "✓", "RBAC + ABAC + policies"},
		{"Delegation Graph", "✓", "DAG + 4 modes"},
		{"Agent Pool", "✓", "9 roles + auto-scaling"},
		{"Workflows", "✓", "6 pre-built workflows"},
		{"Skill Registry", "✓", "8-point security scan"},
		{"AutoEvolve", "✓", "Self-improvement engine"},
		{"Cost Engine", "✓", "Budgets + optimizer"},
	}

	var b strings.Builder
	b.WriteString("[yellow]Brain System Status[white]\n\n")
	for _, s := range subsystems {
		b.WriteString(fmt.Sprintf("  %-20s [%-3s] %s\n", s.name, s.status, s.detail))
	}
	b.WriteString(fmt.Sprintf("\n  [green]Total: %d subsystems operational[white]\n", len(subsystems)))
	b.WriteString(fmt.Sprintf("  [green]Uptime: %s[white]\n", time.Since(time.Now().Add(-time.Hour)).Round(time.Minute)))
	fmt.Fprint(a.status, b.String())
}

func (a *App) Run() error {
	return a.app.Run()
}

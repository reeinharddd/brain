package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

type TestSuiteConfig struct {
	Name    string
	Runner  string
	Path    string
	Timeout time.Duration
}

type TestConfig struct {
	LoggingOutputDir string
	Suites           map[string]TestSuiteConfig
}

type CLIOptions struct {
	Suite       string
	Watch       bool
	OnlyChanged bool
	Debug       bool
	Timeout     time.Duration
}

type runStats struct {
	passed  int
	failed  int
	skipped int
}

type goTestEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

func main() {
	exitCode := run(os.Args[1:])
	os.Exit(exitCode)
}

func run(args []string) int {
	opts, err := parseCLIOptions(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}

	root, err := resolveBrainRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}

	cfg, err := loadTestConfig(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}

	logger, err := NewLogger(root, cfg.LoggingOutputDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	defer logger.Close()

	suiteCfg, ok := cfg.Suites[opts.Suite]
	if !ok {
		logger.Event(LogRecord{Event: "TEST_FAIL", Suite: opts.Suite, Error: "unknown test suite"})
		return 2
	}
	if opts.Timeout > 0 {
		suiteCfg.Timeout = opts.Timeout
	}

	testFiles, err := discoverTestFiles(root, opts.Suite)
	if err != nil {
		logger.Event(LogRecord{Event: "TEST_FAIL", Suite: opts.Suite, Error: err.Error()})
		return 2
	}

	logger.Event(LogRecord{Event: "START", Suite: opts.Suite, TotalTests: len(testFiles), Message: "Loading test configuration"})

	graphCachePath := filepath.Join(root, ".logs", "test-graph.json")
	graph, err := LoadDependencyGraph(graphCachePath)
	if err != nil {
		graph = NewDependencyGraph()
		if scanErr := graph.ScanDependencies(root); scanErr == nil {
			_ = graph.Save(graphCachePath)
		}
	}

	if opts.Debug {
		logger.Event(LogRecord{Event: "DURATION", Suite: opts.Suite, Message: fmt.Sprintf("graph files=%d tests=%d", len(graph.FileToTests), len(graph.TestToFiles))})
	}

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if opts.Watch {
		logger.Event(LogRecord{Event: "DURATION", Suite: opts.Suite, Message: "watch mode placeholder enabled for Phase 1 (single run)"})
	}

	targetPackages := []string{suiteCfg.Path}
	if opts.OnlyChanged {
		changedFiles, changedErr := DetectChangedFiles(root)
		if changedErr == nil && len(changedFiles) > 0 {
			affected := graph.AffectedTests(changedFiles)
			if opts.Debug {
				logger.Event(LogRecord{Event: "DURATION", Suite: opts.Suite, Message: fmt.Sprintf("changed=%d affected_tests=%d", len(changedFiles), len(affected))})
			}
			affectedPackages := packagesFromTests(root, opts.Suite, affected)
			if len(affectedPackages) > 0 {
				targetPackages = affectedPackages
			}
		}
	}

	start := time.Now()
	stats, runErr := runGoTestSuite(ctx, root, suiteCfg, targetPackages, logger)
	duration := time.Since(start)

	logger.Event(LogRecord{
		Event:     "SUMMARY",
		Suite:     opts.Suite,
		Passed:    stats.passed,
		Failed:    stats.failed,
		Skipped:   stats.skipped,
		DurationS: int64(duration.Seconds()),
	})

	if errors.Is(ctx.Err(), context.Canceled) {
		return 130
	}
	if runErr != nil && stats.failed == 0 {
		return 2
	}
	if stats.failed > 0 {
		return 1
	}
	return 0
}

func parseCLIOptions(args []string) (CLIOptions, error) {
	if len(args) > 0 && args[0] == "test" {
		args = args[1:]
	}

	fs := flag.NewFlagSet("testor", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	suiteFlag := fs.String("suite", "daemon", "which test suite to run")
	watch := fs.Bool("watch", false, "watch mode")
	onlyChanged := fs.Bool("onlyChanged", false, "run only changed targets")
	debug := fs.Bool("debug", false, "verbose debug logging")
	timeout := fs.Duration("timeout", 0, "override suite timeout")

	if err := fs.Parse(args); err != nil {
		return CLIOptions{}, err
	}

	suite := strings.TrimSpace(*suiteFlag)
	if fs.NArg() > 0 {
		suite = strings.TrimSpace(fs.Arg(0))
	}
	if suite == "" {
		suite = "daemon"
	}

	return CLIOptions{
		Suite:       suite,
		Watch:       *watch,
		OnlyChanged: *onlyChanged,
		Debug:       *debug,
		Timeout:     *timeout,
	}, nil
}

func resolveBrainRoot() (string, error) {
	if envRoot := strings.TrimSpace(os.Getenv("BRAIN_ROOT")); isBrainRoot(envRoot) {
		return envRoot, nil
	}

	if cwd, err := os.Getwd(); err == nil {
		search := cwd
		for {
			if isBrainRoot(search) {
				return search, nil
			}
			parent := filepath.Dir(search)
			if parent == search {
				break
			}
			search = parent
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	fallback := filepath.Join(home, ".brain")
	if isBrainRoot(fallback) {
		return fallback, nil
	}

	return "", fmt.Errorf("brain root not found")
}

func isBrainRoot(path string) bool {
	if path == "" {
		return false
	}
	manifestPath := filepath.Join(path, "manifest.yml")
	if _, err := os.Stat(manifestPath); err != nil {
		return false
	}
	return true
}

func loadTestConfig(root string) (*TestConfig, error) {
	cfg := &TestConfig{
		LoggingOutputDir: ".logs",
		Suites: map[string]TestSuiteConfig{
			"daemon": {
				Name:    "daemon",
				Runner:  "go_test",
				Path:    "./...",
				Timeout: 60 * time.Second,
			},
			"cli": {
				Name:    "cli",
				Runner:  "go_test",
				Path:    "./...",
				Timeout: 45 * time.Second,
			},
		},
	}

	configPath := filepath.Join(root, "testconfig.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	section := ""
	currentSuite := ""
	lines := strings.Split(string(data), "\n")
	for _, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		if indent == 0 {
			currentSuite = ""
			if strings.HasSuffix(trimmed, ":") {
				section = strings.TrimSuffix(trimmed, ":")
			}
			continue
		}

		if section == "test_suites" && indent == 2 && strings.HasSuffix(trimmed, ":") {
			currentSuite = strings.TrimSuffix(trimmed, ":")
			suite := cfg.Suites[currentSuite]
			suite.Name = currentSuite
			if suite.Runner == "" {
				suite.Runner = "go_test"
			}
			if suite.Path == "" {
				suite.Path = "./..."
			}
			if suite.Timeout == 0 {
				suite.Timeout = 60 * time.Second
			}
			cfg.Suites[currentSuite] = suite
			continue
		}

		if section == "test_suites" && currentSuite != "" && indent >= 4 {
			key, value, ok := splitYAMLKV(trimmed)
			if !ok {
				continue
			}
			suite := cfg.Suites[currentSuite]
			switch key {
			case "runner":
				suite.Runner = cleanYAMLValue(value)
			case "path":
				suite.Path = cleanYAMLValue(value)
			case "timeout":
				timeoutValue := cleanYAMLValue(value)
				if parsed, parseErr := time.ParseDuration(timeoutValue); parseErr == nil {
					suite.Timeout = parsed
				}
			}
			cfg.Suites[currentSuite] = suite
			continue
		}

		if section == "logging" && indent >= 2 {
			key, value, ok := splitYAMLKV(trimmed)
			if !ok {
				continue
			}
			if key == "output_dir" {
				cfg.LoggingOutputDir = cleanYAMLValue(value)
			}
		}
	}

	return cfg, nil
}

func splitYAMLKV(line string) (string, string, bool) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
}

func cleanYAMLValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"")
	value = strings.Trim(value, "'")
	return value
}

func discoverTestFiles(root, suite string) ([]string, error) {
	base := filepath.Join(root, suite)
	if suite == "daemon" {
		base = filepath.Join(root, "daemon")
	} else if suite == "cli" {
		base = filepath.Join(root, "cli")
	}

	files := make([]string, 0)
	err := filepath.WalkDir(base, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "vendor" || strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			files = append(files, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func runGoTestSuite(ctx context.Context, root string, suite TestSuiteConfig, paths []string, logger *Logger) (runStats, error) {
	moduleDir := filepath.Join(root, suite.Name)
	if suite.Name == "daemon" {
		moduleDir = filepath.Join(root, "daemon")
	}
	if suite.Name == "cli" {
		moduleDir = filepath.Join(root, "cli")
	}

	if len(paths) == 0 {
		paths = []string{suite.Path}
	}
	for i := range paths {
		paths[i] = normalizeSuitePath(suite.Name, paths[i])
	}

	args := []string{"test", "-count=1", "-json"}
	args = append(args, paths...)

	logger.Event(LogRecord{Event: "TEST_START", Suite: suite.Name, Message: "Running go test"})

	runCtx := ctx
	cancel := func() {}
	if suite.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, suite.Timeout)
	}
	defer cancel()

	cmd := exec.CommandContext(runCtx, "go", args...)
	cmd.Dir = moduleDir
	output, err := cmd.CombinedOutput()

	stats := runStats{}
	lastOutputByTest := make(map[string]string)
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var evt goTestEvent
		if unmarshalErr := json.Unmarshal([]byte(line), &evt); unmarshalErr != nil {
			continue
		}

		if evt.Action == "output" && evt.Test != "" {
			trimmed := strings.TrimSpace(evt.Output)
			if trimmed != "" {
				lastOutputByTest[evt.Test] = trimmed
			}
			continue
		}

		switch evt.Action {
		case "run":
			if evt.Test != "" {
				logger.Event(LogRecord{Event: "TEST_START", Suite: suite.Name, Name: evt.Test, File: evt.Package})
			}
		case "pass":
			if evt.Test != "" {
				stats.passed++
				logger.Event(LogRecord{Event: "TEST_PASS", Suite: suite.Name, Name: evt.Test, File: evt.Package, DurationMS: int64(evt.Elapsed * 1000)})
			}
		case "fail":
			if evt.Test != "" {
				stats.failed++
				logger.Event(LogRecord{Event: "TEST_FAIL", Suite: suite.Name, Name: evt.Test, File: evt.Package, DurationMS: int64(evt.Elapsed * 1000), Error: lastOutputByTest[evt.Test]})
			}
		case "skip":
			if evt.Test != "" {
				stats.skipped++
				logger.Event(LogRecord{Event: "TEST_SKIP", Suite: suite.Name, Name: evt.Test, File: evt.Package})
			}
		}
	}

	if runCtx.Err() == context.DeadlineExceeded {
		logger.Event(LogRecord{Event: "TEST_FAIL", Suite: suite.Name, Error: "suite timeout exceeded"})
		return stats, runCtx.Err()
	}
	if err != nil && stats.failed == 0 {
		logger.Event(LogRecord{Event: "TEST_FAIL", Suite: suite.Name, Error: strings.TrimSpace(string(output))})
	}

	return stats, err
}

func normalizeSuitePath(suiteName, configuredPath string) string {
	configuredPath = strings.TrimSpace(configuredPath)
	if configuredPath == "" {
		return "./..."
	}

	if suiteName == "daemon" {
		configuredPath = strings.Replace(configuredPath, "./daemon", ".", 1)
	}
	if suiteName == "cli" {
		configuredPath = strings.Replace(configuredPath, "./cli", ".", 1)
	}

	if configuredPath == "." || configuredPath == "./" {
		return "./..."
	}
	return configuredPath
}

func packagesFromTests(root, suite string, testFiles []string) []string {
	if len(testFiles) == 0 {
		return nil
	}

	base := filepath.Join(root, suite)
	if suite == "daemon" {
		base = filepath.Join(root, "daemon")
	}
	if suite == "cli" {
		base = filepath.Join(root, "cli")
	}

	set := make(map[string]struct{})
	for _, testFile := range testFiles {
		dir := filepath.Dir(filepath.FromSlash(testFile))
		rel, err := filepath.Rel(base, dir)
		if err != nil {
			continue
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			set["./"] = struct{}{}
			continue
		}
		set["./"+rel] = struct{}{}
	}

	pkgs := make([]string, 0, len(set))
	for pkg := range set {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)
	return pkgs
}

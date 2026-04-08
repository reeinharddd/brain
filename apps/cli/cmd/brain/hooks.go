package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	coreartifacts "github.com/reeinharrrd/brain/core/artifacts"
)

var (
	envFilePattern          = regexp.MustCompile(`(?i)\.env(\.|$|/)`)
	secretShellPattern      = regexp.MustCompile(`(?i)echo.*(API_KEY|SECRET|PASSWORD|TOKEN)`)
	shellConfigWritePattern = regexp.MustCompile(`(?i)>>?\s*\.(env|bashrc|zshrc|profile)`)
	privateKeyPattern       = regexp.MustCompile(`BEGIN (RSA|EC|DSA|OPENSSH) PRIVATE KEY`)
	awsAccessKeyPattern     = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
	githubTokenPattern      = regexp.MustCompile(`(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36,}`)
)

func HandleHooksCommand(args []string) {
	if len(args) < 3 {
		printHooksHelp()
		return
	}

	switch args[2] {
	case "pre-tool-use":
		if len(args) < 4 {
			printHooksHelp()
			return
		}
		switch args[3] {
		case "block-env-writes":
			os.Exit(runBlockEnvWrites(hookInput(args[4:])))
		default:
			printHooksHelp()
		}
	case "post-tool-use":
		if len(args) < 4 {
			printHooksHelp()
			return
		}
		switch args[3] {
		case "run-linter":
			os.Exit(runPostToolUseLinter())
		default:
			printHooksHelp()
		}
	default:
		printHooksHelp()
	}
}

func printHooksHelp() {
	fmt.Println("Usage: brain hooks <pre-tool-use|post-tool-use> <action>")
	fmt.Println("Actions:")
	fmt.Println("  pre-tool-use block-env-writes")
	fmt.Println("  post-tool-use run-linter")
}

func printMCPServerHelp() {
	fmt.Println("Usage: brain mcp serve [--brain-dir PATH]")
	fmt.Println("Runs the Brain MCP stdio server.")
}

func hookInput(args []string) string {
	for _, key := range []string{"CLAUDE_TOOL_INPUT", "TOOL_INPUT"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}

	if len(args) > 0 {
		return strings.TrimSpace(strings.Join(args, " "))
	}

	return ""
}

func runBlockEnvWrites(input string) int {
	if message, blocked := blockEnvWritesResult(input); blocked {
		fmt.Println(message)
		return 2
	}

	return 0
}

func blockEnvWritesResult(input string) (string, bool) {
	if input == "" {
		return "", false
	}

	switch {
	case envFilePattern.MatchString(input):
		return "[GUARDIAN BLOCKED] Attempt to write to .env file detected.\n  Agents should never modify .env files directly.\n  To add a secret: edit .env manually, or use a secrets manager.", true
	case secretShellPattern.MatchString(input) && shellConfigWritePattern.MatchString(input):
		return "[GUARDIAN BLOCKED] Attempt to write a secret to a shell config file.\n  Manage secrets manually or via a secrets manager.", true
	case privateKeyPattern.MatchString(input):
		return "[GUARDIAN BLOCKED] Private key content detected in tool input.\n  Never let an AI agent handle raw private key material.", true
	case awsAccessKeyPattern.MatchString(input):
		return "[GUARDIAN BLOCKED] AWS Access Key ID detected.\n  Never commit AWS credentials.", true
	case githubTokenPattern.MatchString(input):
		return "[GUARDIAN BLOCKED] GitHub token detected.\n  Never commit GitHub tokens.", true
	default:
		return "", false
	}
}

func runPostToolUseLinter() int {
	file := strings.TrimSpace(os.Getenv("TOOL_OUTPUT_PATH"))
	if file == "" {
		return 0
	}
	if _, err := os.Stat(file); err != nil {
		return 0
	}

	label, command, args, ok := linterSelectionForExt(strings.TrimPrefix(strings.ToLower(filepath.Ext(file)), "."))
	if !ok || !commandAvailable(command) {
		return 0
	}

	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err == nil {
		fmt.Printf("  OK %s: no issues\n", label)
		return 0
	}

	_ = output
	fmt.Printf("  WARN %s: issues found in %s - run: %s %s\n", label, file, command, strings.Join(args, " "))
	return 0
}

func linterForFile(file string) (string, string, []string, bool) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file), "."))

	switch ext {
	case "js", "jsx", "mjs", "cjs", "ts", "tsx":
		if commandAvailable("biome") {
			return "Biome", "biome", []string{"lint", file}, true
		}
		if commandAvailable("eslint") {
			return "ESLint", "eslint", []string{"--no-eslintrc", "--rule", `{"no-unused-vars":"warn"}`, file}, true
		}
	case "py":
		if commandAvailable("ruff") {
			return "Ruff", "ruff", []string{"check", file}, true
		}
	case "sh", "bash":
		if commandAvailable("shellcheck") {
			return "ShellCheck", "shellcheck", []string{file}, true
		}
	}

	return "", "", nil, false
}

func linterSelectionForExt(ext string) (string, string, []string, bool) {
	switch ext {
	case "js", "jsx", "mjs", "cjs", "ts", "tsx":
		return "Biome", "biome", []string{"lint", "FILE"}, true
	case "py":
		return "Ruff", "ruff", []string{"check", "FILE"}, true
	case "sh", "bash":
		return "ShellCheck", "shellcheck", []string{"FILE"}, true
	default:
		return "", "", nil, false
	}
}

func commandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func runHookCommand(args []string) int {
	if len(args) < 2 {
		printHookHelp()
		return 1
	}

	phase := args[0]
	name := args[1]
	var out io.Writer = os.Stdout

	switch phase {
	case "pre-tool-use":
		switch name {
		case "block-env-writes":
			return hookBlockEnvWrites(out)
		case "inject-global-rules":
			return hookInjectGlobalRules(out)
		default:
			fmt.Fprintln(out, "[warn] unknown pre-tool-use hook:", name)
			return 1
		}
	case "post-tool-use":
		switch name {
		case "run-linter":
			return hookRunLinter(out)
		case "auto-invoke-agents":
			return hookAutoInvokeAgents(out)
		case "run-auto-update":
			return hookRunAutoUpdate(out)
		default:
			fmt.Fprintln(out, "[warn] unknown post-tool-use hook:", name)
			return 1
		}
	default:
		fmt.Fprintln(out, "[warn] unknown hook phase:", phase)
		return 1
	}
}

func hookBlockEnvWrites(out io.Writer) int {
	input := hookInputText()
	if input == "" {
		return 0
	}

	blockedPatterns := []struct {
		pattern string
		message string
	}{
		{`\.env(\.|$|/)`, "Attempt to write to a .env file detected."},
		{`BEGIN (RSA|EC|DSA|OPENSSH) PRIVATE KEY`, "Private key content detected in tool input."},
		{`AKIA[0-9A-Z]{16}`, "AWS Access Key ID detected."},
		{`(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36,}`, "GitHub token detected."},
	}

	for _, rule := range blockedPatterns {
		if regexp.MustCompile(rule.pattern).FindString(input) != "" {
			fmt.Fprintln(out, "[blocked]", rule.message)
			if strings.Contains(rule.pattern, `\.env`) {
				fmt.Fprintln(out, "Agents should never modify .env files directly.")
				fmt.Fprintln(out, "Edit the file manually or use a secrets manager.")
			}
			return 2
		}
	}

	if regexp.MustCompile(`echo.*(API_KEY|SECRET|PASSWORD|TOKEN)`).FindString(input) != "" &&
		regexp.MustCompile(`>>?\s*\.(env|bashrc|zshrc|profile)`).FindString(input) != "" {
		fmt.Fprintln(out, "[blocked] Attempt to write a secret to a shell config file.")
		fmt.Fprintln(out, "Manage secrets manually or via a secrets manager.")
		return 2
	}

	return 0
}

func hookInjectGlobalRules(out io.Writer) int {
	brainDir := resolveBrainRoot()
	locator := coreartifacts.NewLocator(brainDir)
	canonical := locator.DomainFile("rules", "canonical.md")
	canonicalInfo, err := os.Stat(canonical)
	if err != nil {
		fmt.Fprintln(out, "[warn] canonical rules not available:", err)
		return 0
	}

	adapters := []string{
		filepath.Join(locator.DomainDir("adapters"), "claude-code", "CLAUDE.md"),
		filepath.Join(locator.DomainDir("adapters"), "cursor", ".cursorrules"),
		filepath.Join(locator.DomainDir("adapters"), "windsurf", ".windsurfrules"),
	}

	outdated := false
	for _, adapter := range adapters {
		info, err := os.Stat(adapter)
		if err != nil {
			fmt.Fprintln(out, "[warn] missing adapter:", adapter)
			outdated = true
			continue
		}
		if canonicalInfo.ModTime().After(info.ModTime()) {
			fmt.Fprintln(out, "[warn] adapter outdated:", filepath.Base(adapter))
			outdated = true
		}
	}

	if outdated {
		fmt.Fprintln(out, "Run: brain sync")
	} else {
		fmt.Fprintln(out, "[ok] adapters are up to date")
	}

	return 0
}

func hookRunLinter(out io.Writer) int {
	file := hookOutputPath()
	if file == "" {
		return 0
	}
	if _, err := os.Stat(file); err != nil {
		return 0
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(file)), ".")
	switch ext {
	case "js", "jsx", "mjs", "cjs", "ts", "tsx":
		if commandExists("biome") {
			runLinterCommand(out, "biome", []string{"lint", file}, "Biome")
			return 0
		}
		if commandExists("eslint") {
			runLinterCommand(out, "eslint", []string{"--no-eslintrc", "--rule", `{"no-unused-vars": "warn"}`, file}, "ESLint")
			return 0
		}
	case "py":
		if commandExists("ruff") {
			runLinterCommand(out, "ruff", []string{"check", file}, "Ruff")
			return 0
		}
	case "sh", "bash":
		if commandExists("shellcheck") {
			runLinterCommand(out, "shellcheck", []string{file}, "ShellCheck")
			return 0
		}
	}

	return 0
}

func hookAutoInvokeAgents(out io.Writer) int {
	toolName := strings.TrimSpace(firstNonEmpty(os.Getenv("CLAUDE_TOOL_NAME"), os.Getenv("TOOL_NAME")))
	if toolName == "" {
		return 0
	}

	switch toolName {
	case "write_file", "edit_file", "replace_file_content", "multi_replace_file_content":
		fmt.Fprintln(out, "[suggest] Task involves file writing. Consider consulting reviewer.")
	case "run_command", "send_command_input":
		fmt.Fprintln(out, "[suggest] Shell execution detected. Consider consulting guardian.")
	case "grep_search", "find_by_name", "search_web":
		fmt.Fprintln(out, "[suggest] Research activity detected. researcher may have more context.")
	case "task_boundary":
		fmt.Fprintln(out, "[suggest] Planning in progress. Ensure planner has defined the roadmap.")
	}

	return 0
}

func hookRunAutoUpdate(out io.Writer) int {
	paths, err := gitChangedPaths()
	if err != nil {
		fmt.Fprintln(out, "[warn] unable to inspect git diff:", err)
		return 0
	}

	needsSync := false
	trackedDomains := []string{
		"skills",
		"agents",
		"commands",
		"rules",
		"mcps",
		"providers",
		"adapters",
	}

	for _, path := range paths {
		for _, domain := range trackedDomains {
			if coreartifacts.PathInDomain(path, domain) {
				needsSync = true
				break
			}
		}

		if needsSync {
			break
		}
	}

	if !needsSync {
		fmt.Fprintln(out, "[ok] No skill or agent changes detected.")
		return 0
	}

	fmt.Fprintln(out, "[info] Skill or agent changes detected. Triggering brain sync.")
	triggerSync(false)
	return 0
}

func gitChangedPaths() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths, nil
}

func runLinterCommand(out io.Writer, command string, args []string, label string) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err == nil {
		fmt.Fprintln(out, "[ok]", label+": no issues")
		return
	}

	if len(output) > 0 {
		fmt.Fprintln(out, "[warn]", label+" issues found")
		fmt.Fprintln(out, strings.TrimSpace(string(output)))
		return
	}

	fmt.Fprintln(out, "[warn]", label+" issues found")
}

func hookInputText() string {
	if value := strings.TrimSpace(os.Getenv("TOOL_INPUT")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("CLAUDE_TOOL_INPUT")); value != "" {
		return value
	}
	if len(os.Args) > 2 {
		return strings.Join(os.Args[2:], " ")
	}
	return ""
}

func hookOutputPath() string {
	return strings.TrimSpace(firstNonEmpty(os.Getenv("TOOL_OUTPUT_PATH"), os.Getenv("CLAUDE_TOOL_OUTPUT_PATH")))
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func printHookHelp() {
	fmt.Fprintln(os.Stdout, "Usage: brain hook <pre-tool-use|post-tool-use> <hook-name>")
}

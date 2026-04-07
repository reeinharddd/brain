package main

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const daemonModuleImport = "github.com/reeinharrrd/brain/daemon"

type DependencyGraph struct {
	FileToTests map[string][]string `json:"file_to_tests"`
	TestToFiles map[string][]string `json:"test_to_files"`
	BuiltAt     string              `json:"built_at"`
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		FileToTests: make(map[string][]string),
		TestToFiles: make(map[string][]string),
	}
}

func (g *DependencyGraph) ScanDependencies(root string) error {
	fset := token.NewFileSet()
	sourcesByDir := make(map[string][]string)
	testFiles := make([]string, 0)

	daemonRoot := filepath.Join(root, "daemon")
	err := filepath.WalkDir(daemonRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "vendor" || strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		norm := normalizePath(path)
		dir := filepath.Dir(norm)
		if strings.HasSuffix(norm, "_test.go") {
			testFiles = append(testFiles, norm)
			g.FileToTests[norm] = appendUnique(g.FileToTests[norm], norm)
			return nil
		}
		sourcesByDir[dir] = append(sourcesByDir[dir], norm)
		return nil
	})
	if err != nil {
		return err
	}

	for _, files := range sourcesByDir {
		sort.Strings(files)
	}

	for _, testFile := range testFiles {
		deps := make([]string, 0)
		deps = append(deps, sourcesByDir[filepath.Dir(testFile)]...)

		astFile, parseErr := parser.ParseFile(fset, testFile, nil, parser.ImportsOnly)
		if parseErr == nil {
			for _, imp := range astFile.Imports {
				importPath := strings.Trim(imp.Path.Value, "\"")
				if !strings.HasPrefix(importPath, daemonModuleImport) {
					continue
				}

				relImport := strings.TrimPrefix(importPath, daemonModuleImport)
				relImport = strings.TrimPrefix(relImport, "/")
				importDir := normalizePath(filepath.Join(daemonRoot, relImport))
				deps = append(deps, sourcesByDir[importDir]...)
			}
		}

		deps = dedupeSortedStrings(deps)
		g.TestToFiles[testFile] = deps
		for _, dep := range deps {
			g.FileToTests[dep] = appendUnique(g.FileToTests[dep], testFile)
		}
	}

	for key := range g.FileToTests {
		g.FileToTests[key] = dedupeSortedStrings(g.FileToTests[key])
	}
	for key := range g.TestToFiles {
		g.TestToFiles[key] = dedupeSortedStrings(g.TestToFiles[key])
	}

	g.BuiltAt = time.Now().UTC().Format(time.RFC3339)
	return nil
}

func (g *DependencyGraph) AffectedTests(changedFiles []string) []string {
	affected := make([]string, 0)
	for _, file := range changedFiles {
		norm := normalizePath(file)
		affected = append(affected, g.FileToTests[norm]...)
	}
	return dedupeSortedStrings(affected)
}

func (g *DependencyGraph) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadDependencyGraph(path string) (*DependencyGraph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var graph DependencyGraph
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, err
	}
	if graph.FileToTests == nil {
		graph.FileToTests = make(map[string][]string)
	}
	if graph.TestToFiles == nil {
		graph.TestToFiles = make(map[string][]string)
	}
	return &graph, nil
}

func DetectChangedFiles(root string) ([]string, error) {
	changed := make([]string, 0)

	diffCmd := exec.Command("git", "-C", root, "diff", "--name-only", "HEAD")
	diffOut, diffErr := diffCmd.Output()
	if diffErr == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(diffOut)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				changed = append(changed, normalizePath(filepath.Join(root, line)))
			}
		}
	}

	untrackedCmd := exec.Command("git", "-C", root, "ls-files", "--others", "--exclude-standard")
	untrackedOut, untrackedErr := untrackedCmd.Output()
	if untrackedErr == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(untrackedOut)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				changed = append(changed, normalizePath(filepath.Join(root, line)))
			}
		}
	}

	if diffErr != nil && untrackedErr != nil {
		return nil, diffErr
	}

	return dedupeSortedStrings(changed), nil
}

func dedupeSortedStrings(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, item := range input {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func appendUnique(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func normalizePath(path string) string {
	clean := filepath.Clean(path)
	return filepath.ToSlash(clean)
}

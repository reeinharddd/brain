package runtime

import (
	"os"
	"path/filepath"
	"strings"
)

const BrainRootEnv = "BRAIN_ROOT"

func ConfigRootFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "brain", "root")
}

func SaveConfiguredRoot(root string) error {
	conf := ConfigRootFilePath()
	if err := os.MkdirAll(filepath.Dir(conf), 0755); err != nil {
		return err
	}
	return os.WriteFile(conf, []byte(root+"\n"), 0644)
}

func ReadConfiguredRoot() string {
	b, err := os.ReadFile(ConfigRootFilePath())
	if err != nil {
		return ""
	}
	root := strings.TrimSpace(string(b))
	if !IsBrainRoot(root) {
		return ""
	}
	return root
}

func IsBrainRoot(path string) bool {
	if path == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "manifest.yml")); err != nil {
		return false
	}
	requiredSets := [][][]string{
		{
			{"apps", "cli", "cmd", "brain", "main.go"},
			{"apps", "daemon", "cmd", "braind", "main.go"},
		},
		{
			{"cli", "cmd", "brain", "main.go"},
			{"daemon", "cmd", "braind", "main.go"},
		},
	}
	for _, checks := range requiredSets {
		ok := true
		for _, parts := range checks {
			if _, err := os.Stat(filepath.Join(append([]string{path}, parts...)...)); err != nil {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func ResolveBrainRoot() string {
	if envRoot := strings.TrimSpace(os.Getenv(BrainRootEnv)); IsBrainRoot(envRoot) {
		return envRoot
	}

	if cwd, err := os.Getwd(); err == nil {
		search := cwd
		for {
			if IsBrainRoot(search) {
				return search
			}
			parent := filepath.Dir(search)
			if parent == search {
				break
			}
			search = parent
		}
	}

	if configured := ReadConfiguredRoot(); configured != "" {
		return configured
	}

	home, _ := os.UserHomeDir()
	fallback := filepath.Join(home, ".brain")
	if IsBrainRoot(fallback) {
		return fallback
	}
	return fallback
}

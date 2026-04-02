package manifest

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Domain struct {
	Source  string `yaml:"source"`
	Enabled bool   `yaml:"enabled"`
}

type Target struct {
	Enabled         bool              `yaml:"enabled"`
	Type            string            `yaml:"type"`
	OutputDirs      map[string]string `yaml:"output_dirs"`
	ManagedSections bool              `yaml:"managed_sections"`
}

type Settings struct {
	BackupEnabled       bool   `yaml:"backup_enabled"`
	BackupDir           string `yaml:"backup_dir"`
	BackupRetentionDays int    `yaml:"backup_retention_days"`
	DryRunByDefault     bool   `yaml:"dry_run_by_default"`
}

type AppManifest struct {
	Version  string            `yaml:"version"`
	Settings Settings          `yaml:"settings"`
	Domains  map[string]Domain `yaml:"domains"`
	Targets  map[string]Target `yaml:"targets"`
}

func Parse(path string) (*AppManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m AppManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

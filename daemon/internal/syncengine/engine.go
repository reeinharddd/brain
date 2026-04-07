package syncengine

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/reeinharrrd/brain/daemon/internal/manager"
	"github.com/reeinharrrd/brain/daemon/internal/manifest"
	"github.com/reeinharrrd/brain/daemon/internal/syncengine/workers"
)

type SyncEngine struct {
	logger         chan<- string
	manifest       *manifest.AppManifest
	skillsRegistry *manager.SkillsRegistry
}

type SyncOptions struct {
	DryRun bool
}

func NewSyncEngine(m *manifest.AppManifest, logger chan<- string, skills *manager.SkillsRegistry) *SyncEngine {
	return &SyncEngine{
		logger:         logger,
		manifest:       m,
		skillsRegistry: skills,
	}
}

func (s *SyncEngine) expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func (s *SyncEngine) RunSync() error {
	return s.RunSyncWithOptions(SyncOptions{})
}

func (s *SyncEngine) RunSyncWithOptions(opts SyncOptions) error {
	s.logger <- "[SyncEngine] Initializing Native Synchronization..."

	dryRun := opts.DryRun || s.manifest.Settings.DryRunByDefault
	if dryRun {
		s.logger <- "[SyncEngine] Dry run enabled - no files will be written"
	}

	if s.manifest.Settings.BackupEnabled && !dryRun {
		if err := s.createBackup(); err != nil {
			s.logger <- fmt.Sprintf("[SyncEngine] Backup Failed: %v", err)
			return err
		}
	}

	rulesWorker := &workers.RulesWorker{}
	mcpWorker := &workers.MCPWorker{}
	skillsWorker := &workers.SkillsWorker{}
	agentsWorker := &workers.AgentsWorker{}

	brainDir := s.expandHome("~/.brain")

	// Map output config keys to workers based on known configurations
	for name, target := range s.manifest.Targets {
		if !target.Enabled {
			s.logger <- fmt.Sprintf("[SyncEngine] Skipping target %s (Disabled)", name)
			continue
		}
		s.logger <- fmt.Sprintf("[SyncEngine] Syncing target: %s", name)

		for key, outDirRaw := range target.OutputDirs {
			outDir := s.expandHome(outDirRaw)

			if dryRun {
				s.logger <- fmt.Sprintf("[SyncEngine] [DRY-RUN] -> Would deploy %s to %s", key, outDir)
				continue
			}

			s.logger <- fmt.Sprintf("[SyncEngine] -> Deploying %s to %s", key, outDir)

			switch key {
			case "rules", "instructions":
				if domConfig, exists := s.manifest.Domains["rules"]; exists && domConfig.Enabled {
					source := filepath.Join(brainDir, domConfig.Source)
					if err := rulesWorker.Sync(source, outDir, s.logger); err != nil {
						return fmt.Errorf("rules sync failed for target %s: %w", name, err)
					}
				}
			case "mcp":
				if domConfig, exists := s.manifest.Domains["mcp"]; exists && domConfig.Enabled {
					source := filepath.Join(brainDir, domConfig.Source)
					if err := mcpWorker.Sync(source, outDir, s.logger); err != nil {
						return fmt.Errorf("mcp sync failed for target %s: %w", name, err)
					}
				}
			case "skills":
				if domConfig, exists := s.manifest.Domains["skills"]; exists && domConfig.Enabled {
					// Pass lambda to get catalog from skills registry
					getCatalogFunc := func() []*workers.CatalogItem {
						items := s.skillsRegistry.GetAll(context.Background())
						// Convert manager.CatalogItem to workers.CatalogItem
						result := make([]*workers.CatalogItem, len(items))
						for i, item := range items {
							result[i] = &workers.CatalogItem{
								ID:          item.ID,
								Name:        item.Name,
								Kind:        item.Kind,
								Scope:       item.Scope,
								Description: item.Description,
								Tags:        item.Tags,
								Path:        item.Path,
								Version:     item.Version,
								Maintained:  item.Maintained,
								Source:      item.Source,
								Type:        item.Type,
								File:        item.File,
								SyncTo:      item.SyncTo,
								Requires:    item.Requires,
								Category:    item.Category,
							}
						}
						return result
					}
					
					if err := skillsWorker.Sync(brainDir, outDir, s.logger, getCatalogFunc); err != nil {
						return fmt.Errorf("skills sync failed for target %s: %w", name, err)
					}
				}
			case "agents":
				if domConfig, exists := s.manifest.Domains["agents"]; exists && domConfig.Enabled {
					source := filepath.Join(brainDir, domConfig.Source)
					if err := agentsWorker.Sync(source, outDir, s.logger); err != nil {
						return fmt.Errorf("agents sync failed for target %s: %w", name, err)
					}
				}
			default:
				s.logger <- fmt.Sprintf("[SyncEngine] Warning: No native worker available for output key: %s (in target %s)", key, name)
			}
		}
	}

	s.logger <- "[SyncEngine] Native Synchronization Complete."
	return nil
}

func (s *SyncEngine) createBackup() error {
	backupDir := s.expandHome(s.manifest.Settings.BackupDir)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("brain_backup_%s.tar.gz", timestamp))

	s.logger <- fmt.Sprintf("[SyncEngine] Creating target configurations backup at %s", backupFile)

	file, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	defer file.Close()
	gw := gzip.NewWriter(file)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, t := range s.manifest.Targets {
		if !t.Enabled {
			continue
		}
		for _, dirRaw := range t.OutputDirs {
			dir := s.expandHome(dirRaw)
			if stat, err := os.Stat(dir); err == nil && !stat.IsDir() {
				s.backupFileToTar(tw, dir)
			}
		}
	}
	return nil
}

func (s *SyncEngine) backupFileToTar(tw *tar.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(stat, stat.Name())
	if err != nil {
		return err
	}
	header.Name = filepath.Base(path)

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}

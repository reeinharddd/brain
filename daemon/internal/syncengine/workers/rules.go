package workers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RulesWorker struct{}

func (w *RulesWorker) Sync(sourceDirPath string, outPath string, logger chan<- string) error {
	baseRuleFile := sourceDirPath
	modulesDir := filepath.Join(filepath.Dir(sourceDirPath), "modules")

	if filepath.Ext(outPath) == "" {
		outPath = filepath.Join(outPath, "brain.instructions.md")
	}

	logger <- fmt.Sprintf("[RulesWorker] Compiling rules into %s", outPath)

	var buffer bytes.Buffer

	canonicalContent, err := ioutil.ReadFile(baseRuleFile)
	if err != nil {
		return fmt.Errorf("failed to read canonical rules: %v", err)
	}

	header := fmt.Sprintf("<!-- AUTO-GENERATED  DO NOT EDIT DIRECTLY -->\n<!-- Generated on %s  Source: %s + modules/ -->\n\n", time.Now().Format("2006-01-02 15:04:05"), baseRuleFile)
	buffer.WriteString(header)
	buffer.WriteString("# VSCode Copilot Instructions\n\n")
	buffer.Write(canonicalContent)

	files, err := ioutil.ReadDir(modulesDir)
	if err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
				modPath := filepath.Join(modulesDir, f.Name())
				modContent, modErr := ioutil.ReadFile(modPath)
				if modErr == nil {
					buffer.WriteString("\n\n")
					buffer.Write(modContent)
				}
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("failed to create target dir: %v", err)
	}

	if err := ioutil.WriteFile(outPath, buffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write compiled rules: %v", err)
	}

	logger <- fmt.Sprintf("[RulesWorker] Successfully wrote compiled rules to %s", outPath)
	return nil
}
